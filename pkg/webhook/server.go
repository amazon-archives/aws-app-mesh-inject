package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awslabs/aws-app-mesh-inject/pkg/config"
	"github.com/awslabs/aws-app-mesh-inject/pkg/patch"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/apis/apps"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNoUID                  = errors.New("No UID from request")
	ErrNoPorts                = errors.New("No ports specified for injection, doing nothing")
	ErrNoName                 = errors.New("No VirtualNode name specified for injection, doing nothing")
	ErrNoObject               = errors.New("No Object passed to mutate")
	portsAnnotation           = "appmesh.k8s.aws/ports"
	virtualNodeNameAnnotation = "appmesh.k8s.aws/virtualNode"
	sidecarInjectAnnotation   = "appmesh.k8s.aws/sidecarInjectorWebhook"
)

type Server struct {
	Config      config.Config
	KubeClient  kubernetes.Interface
	KubeDecoder runtime.Decoder
}

// ListenAndServe starts the mutating webhook HTTP server
func (s *Server) ListenAndServe(enableTLS bool, timeout time.Duration, stopCh <-chan struct{}) {
	mux := http.DefaultServeMux

	// register handlers
	mux.HandleFunc("/healthz", s.healthHandler)
	mux.HandleFunc("/", s.injectHandler)

	// init server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", s.Config.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  15 * time.Second,
	}

	// load TLS cert from disk
	if enableTLS {
		cert, err := tls.LoadX509KeyPair(s.Config.TlsCert, s.Config.TlsKey)
		if err != nil {
			log.Panicf("TLS cert loading failed %s", err.Error())
		}
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	log.Infof("Starting HTTP server on port %v", s.Config.Port)

	// run server in background
	go func() {
		if enableTLS {
			if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
				log.Fatalf("HTTP server crashed %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("HTTP server crashed %v", err)
			}
		}
	}()

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// try graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server graceful shutdown failed %v", err)
	} else {
		log.Info("HTTP server stopped")
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) injectHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var body []byte
	var admissionResponse *v1beta1.AdmissionResponse

	// validate content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	defer r.Body.Close()

	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	receivedAdmissionReview := v1beta1.AdmissionReview{}
	returnedAdmissionReview := v1beta1.AdmissionReview{}

	// decode pod spec and patch it
	if err = s.decode(body, &receivedAdmissionReview); err != nil {
		log.Error(err)
		admissionResponse = admissionResponseError(err)
	} else if err = validateRequest(receivedAdmissionReview); err != nil {
		log.Error(err)
		admissionResponse = admissionResponseError(err)
	} else {
		admissionResponse = s.mutate(receivedAdmissionReview)
		admissionResponse.UID = receivedAdmissionReview.Request.UID
	}

	returnedAdmissionReview.Response = admissionResponse
	resp, err := json.Marshal(returnedAdmissionReview)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Error(err)
	}
}

func (s *Server) mutate(receivedAdmissionReview v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	var ports string
	var name string
	admissionResponse := v1beta1.AdmissionResponse{
		Allowed: true,
	}
	pod := corev1.Pod{}

	if err := json.Unmarshal(receivedAdmissionReview.Request.Object.Raw, &pod); err != nil {
		log.Error(err)
		return admissionResponseError(err)
	}

	// If sidecar injection is disabled in the annotation, we skip mutating
	switch strings.ToLower(pod.ObjectMeta.Annotations[sidecarInjectAnnotation]) {
	case "disabled":
		log.Info("Sidecar inject is disabled. Skipping mutating")
		return &admissionResponse
	}

	// set ports
	if v, ok := pod.ObjectMeta.Annotations[portsAnnotation]; ok {
		ports = v
	} else {
		// if ports isn't specified in the pod annotation, use the container ports from the pod spec.
		// https://github.com/awslabs/aws-app-mesh-inject/issues/2
		portArray := getPortsFromContainers(pod.Spec.Containers)
		if len(portArray) == 0 {
			log.Info(ErrNoPorts)
			return &admissionResponse
		}
		ports = strings.Join(portArray, ",")
	}

	// set virtual node name
	if v, ok := pod.ObjectMeta.Annotations[virtualNodeNameAnnotation]; ok {
		name = v
	} else {
		// if virtual router name isn't specified in the pod annotation, use the controller owner name instead.
		// https://github.com/awslabs/aws-app-mesh-inject/issues/4
		if controllerName := s.getControllerNameForPod(pod, receivedAdmissionReview.Request.Namespace); controllerName != nil {
			name = *controllerName
		} else {
			log.Info(ErrNoName)
			return &admissionResponse
		}
	}

	// patch pod spec
	admissionResponse.Patch = patch.GetPatch(
		len(pod.Spec.InitContainers),
		len(pod.Spec.Containers),
		len(pod.Spec.ImagePullSecrets),
		s.Config.Name,
		s.Config.Region,
		name,
		ports,
		s.Config.LogLevel,
		s.Config.EcrSecret,
	)
	log.Infof("Patch %v", string(admissionResponse.Patch))
	pt := v1beta1.PatchTypeJSONPatch
	admissionResponse.PatchType = &pt

	return &admissionResponse
}

func (s *Server) decode(b []byte, o runtime.Object) error {
	_, _, err := s.KubeDecoder.Decode(b, nil, o)
	return err
}

func admissionResponseError(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func validateRequest(receivedAdmissionReview v1beta1.AdmissionReview) error {
	if receivedAdmissionReview.Request == nil {
		fmt.Println(receivedAdmissionReview)
		return ErrNoObject
	}
	if receivedAdmissionReview.Request.UID == "" {
		return ErrNoUID
	}
	return nil
}

// get the name of the controller that created the pod.
func (s *Server) getControllerNameForPod(pod corev1.Pod, namespace string) *string {
	controllerRef := metav1.GetControllerOf(pod.GetObjectMeta())
	if controllerRef == nil {
		// An orphan
		return nil
	}

	if controllerRef.Kind != apps.SchemeGroupVersion.WithKind("ReplicaSet").Kind {
		// The pod is not owned by a replica set. Return the controller name directly
		return &controllerRef.Name
	}
	rs, err := s.KubeClient.AppsV1().ReplicaSets(namespace).Get(controllerRef.Name, metav1.GetOptions{})
	if err != nil || rs.UID != controllerRef.UID {
		log.Errorf("Cannot get replicaset %q for pod %q: %v", controllerRef.Name, pod.Name, err)
		return nil
	}

	// Now find the Controller that owns that ReplicaSet.
	parentControllerRef := metav1.GetControllerOf(rs)
	if parentControllerRef == nil {
		// The replica set created the pod
		return &controllerRef.Name
	}
	return &parentControllerRef.Name
}

// get all the ports from containers
func getPortsFromContainers(containers []corev1.Container) []string {
	parts := make([]string, 0)
	for _, container := range containers {
		parts = append(parts, getPortsForContainer(container)...)
	}

	return parts
}

// get all the ports for that container
func getPortsForContainer(container corev1.Container) []string {
	parts := make([]string, 0)
	for _, p := range container.Ports {
		parts = append(parts, strconv.Itoa(int(p.ContainerPort)))
	}
	return parts
}
