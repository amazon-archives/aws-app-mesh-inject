package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-app-mesh-inject/pkg/config"
	"github.com/aws/aws-app-mesh-inject/pkg/patch"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/apps"
)

var (
	ErrNoUID    = errors.New("No UID from request")
	ErrNoPorts  = errors.New("No ports specified for injection, doing nothing")
	ErrNoName   = errors.New("No VirtualNode name specified for injection, doing nothing")
	ErrNoObject = errors.New("No Object passed to mutate")
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
			klog.Fatalf("TLS cert loading failed %s", err.Error())
		}
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	klog.Infof("Starting HTTP server on port %v", s.Config.Port)

	// run server in background
	go func() {
		if enableTLS {
			if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
				klog.Fatalf("HTTP server crashed %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				klog.Fatalf("HTTP server crashed %v", err)
			}
		}
	}()

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// try graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		klog.Errorf("HTTP server graceful shutdown failed %v", err)
	} else {
		klog.Info("HTTP server stopped")
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
		klog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	receivedAdmissionReview := v1beta1.AdmissionReview{}
	returnedAdmissionReview := v1beta1.AdmissionReview{}

	// decode pod spec and patch it
	if err = s.decode(body, &receivedAdmissionReview); err != nil {
		klog.Error(err)
		admissionResponse = admissionResponseError(err)
	} else if err = validateRequest(receivedAdmissionReview); err != nil {
		klog.Error(err)
		admissionResponse = admissionResponseError(err)
	} else {
		admissionResponse = s.mutate(receivedAdmissionReview)
		admissionResponse.UID = receivedAdmissionReview.Request.UID
	}

	returnedAdmissionReview.Response = admissionResponse
	resp, err := json.Marshal(returnedAdmissionReview)
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		klog.Error(err)
	}
}

func (s *Server) mutate(receivedAdmissionReview v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	var ports string
	var egressIgnoredPorts string
	var name string
	admissionResponse := v1beta1.AdmissionResponse{
		Allowed: true,
	}
	pod := corev1.Pod{}

	if err := json.Unmarshal(receivedAdmissionReview.Request.Object.Raw, &pod); err != nil {
		klog.Error(err)
		return admissionResponseError(err)
	}

	shouldInject, err := s.shouldInject(pod)
	if err != nil {
		return admissionResponseError(err)
	}

	if shouldInject == false {
		klog.Info("Sidecar inject is disabled. Skipping mutation")
		return &admissionResponse
	}

	// set mesh name
	meshName := s.Config.MeshName
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshMeshNameAnnotation]; ok {
		meshName = v
	}

	// set egress ignored ports
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshEgressIgnoredPortsAnnotation]; ok {
		egressIgnoredPorts = v
	} else {
		egressIgnoredPorts = "22"
	}

	// set ports
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshPortsAnnotation]; ok {
		ports = v
	} else {
		// if ports isn't specified in the pod annotation, use the container ports from the pod spec.
		// https://github.com/aws/aws-app-mesh-inject/issues/2
		portArray := getPortsFromContainers(pod.Spec.Containers)
		if len(portArray) == 0 {
			klog.Info(ErrNoPorts)
			return &admissionResponse
		}
		ports = strings.Join(portArray, ",")
	}

	// set virtual node name
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshVirtualNodeNameAnnotation]; ok {
		name = v
	} else {
		// if virtual router name isn't specified in the pod annotation, use the controller owner name instead.
		// https://github.com/aws/aws-app-mesh-inject/issues/4
		if controllerName := s.getControllerNameForPod(pod, receivedAdmissionReview.Request.Namespace); controllerName != nil {
			name = fmt.Sprintf("%s-%s", *controllerName, receivedAdmissionReview.Request.Namespace)
		} else {
			klog.Info(ErrNoName)
			return &admissionResponse
		}
	}

	// set preview channel enabled
	var preview string
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshPreviewAnnotation]; ok {
		if v == "true" {
			preview = "1"
		} else {
			preview = "0"
		}
	} else {
		if s.Config.Preview {
			preview = "1"
		} else {
			preview = "0"
		}
	}

	// set cpu-request
	var cpuRequest string
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshCpuRequestAnnotation]; ok {
		cpuRequest = v
	} else {
		cpuRequest = s.Config.SidecarCpu
	}

	// set memory-request
	var memoryRequest string
	if v, ok := pod.ObjectMeta.Annotations[config.AppMeshMemoryRequestAnnotation]; ok {
		memoryRequest = v
	} else {
		memoryRequest = s.Config.SidecarMemory
	}

	klog.Infof("Patching pod %v", pod.ObjectMeta)

	// patch pod spec
	podPatch, err := patch.GeneratePatch(patch.Meta{
		PodAnnotations:        pod.ObjectMeta.Annotations,
		HasImagePullSecret:    s.Config.EcrSecret,
		AppendImagePullSecret: len(pod.Spec.ImagePullSecrets) > 0,
		AppendInit:            len(pod.Spec.InitContainers) > 0,
		AppendSidecar:         len(pod.Spec.Containers) > 0,
		Init: patch.InitMeta{
			Ports:              ports,
			EgressIgnoredPorts: egressIgnoredPorts,
			ContainerImage:     s.Config.InitImage,
			IgnoredIPs:         s.Config.IgnoredIPs,
			MemoryRequests:     s.Config.SidecarMemory,
			CpuRequests:        s.Config.SidecarCpu,
		},
		Sidecar: patch.SidecarMeta{
			VirtualNodeName:      name,
			ContainerImage:       s.Config.SidecarImage,
			LogLevel:             s.Config.LogLevel,
			Region:               s.Config.Region,
			Preview:              preview,
			MeshName:             meshName,
			MemoryRequests:       memoryRequest,
			CpuRequests:          cpuRequest,
			EnableJaegerTracing:  s.Config.EnableJaegerTracing,
			JaegerAddress:        s.Config.JaegerAddress,
			JaegerPort:           s.Config.JaegerPort,
			EnableDatadogTracing: s.Config.EnableDatadogTracing,
			DatadogAddress:       s.Config.DatadogAddress,
			DatadogPort:          s.Config.DatadogPort,
			InjectXraySidecar:    s.Config.InjectXraySidecar,
			EnableStatsTags:      s.Config.EnableStatsTags,
			EnableStatsD:         s.Config.EnableStatsD,
		},
	})
	if err != nil {
		klog.Error(err)
		return admissionResponseError(err)
	}

	admissionResponse.Patch = podPatch
	pt := v1beta1.PatchTypeJSONPatch
	admissionResponse.PatchType = &pt

	return &admissionResponse
}

// Determines if the sidecars should be injected or not
func (s *Server) shouldInject(pod corev1.Pod) (bool, error) {
	// If sidecar injection is explicitely set in the annotation, honor it
	sidecarInjection := strings.ToLower(pod.ObjectMeta.Annotations[config.AppMeshSidecarInjectAnnotation])
	if sidecarInjection != "" {
		inject, err := s.convertAnnotation(sidecarInjection)
		if err != nil {
			klog.Warning(fmt.Sprintf("Error converting annotation: %v for pod %v/%v. Using default injection value", err, pod.GetNamespace(), pod.GetName()))
			inject = s.Config.InjectDefault
		}
		klog.Info(fmt.Sprintf("Sidecar inject is %v in %v for pod %v/%v", inject, config.AppMeshSidecarInjectAnnotation, pod.GetNamespace(), pod.GetName()))
		return inject, nil
	}

	// Otherwise use the default value
	return s.Config.InjectDefault, nil
}

// Converts the injection annotation string to a boolean
func (s *Server) convertAnnotation(annotation string) (bool, error) {
	switch strings.ToLower(annotation) {
	case "enabled":
		return true, nil
	case "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("Invalid %v annotation: %v", config.AppMeshSidecarInjectAnnotation, annotation)
	}
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
		klog.Errorf("Cannot get replicaset %q for pod %q: %v", controllerRef.Name, pod.Name, err)
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
