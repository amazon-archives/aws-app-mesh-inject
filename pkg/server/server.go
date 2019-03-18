package server

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/awslabs/aws-app-mesh-inject/pkg/config"
	"github.com/awslabs/aws-app-mesh-inject/pkg/patch"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/apis/apps"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	corev1.AddToScheme(scheme)
	admissionregistrationv1beta1.AddToScheme(scheme)
	flag.StringVar(&tlscert, "tlscert", "/etc/webhook/certs/cert.pem", "Location of TLS Cert file.")
	flag.StringVar(&tlskey, "tlskey", "/etc/webhook/certs/key.pem", "Location of TLS key file.")
}

var (
	scheme                    = runtime.NewScheme()
	codecs                    = serializer.NewCodecFactory(scheme)
	tlscert, tlskey           string
	healthResponse            = []byte("200 - Healthy")
	wrongContentResponse      = []byte("415 - Wrong Content Type")
	ErrNoUID                  = errors.New("No UID from request")
	ErrNoPorts                = errors.New("No ports specified for injection, doing nothing")
	ErrNoName                 = errors.New("No VirtualNode name specified for injection, doing nothing")
	ErrNoObject               = errors.New("No Object passed to mutate")
	portsAnnotation           = "appmesh.k8s.aws/ports"
	virtualNodeNameAnnotation = "appmesh.k8s.aws/virtualNode"
	sidecarInjectAnnotation   = "appmesh.k8s.aws/sidecarInjectorWebhook"

	kubeconfig, _ = rest.InClusterConfig()
	clientset, _  = kubernetes.NewForConfig(kubeconfig)
)

func admissionResponseError(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

type AppMeshHandler struct {
	config.Config
}

func (ah AppMeshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var admissionResponse *v1beta1.AdmissionResponse
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	log.Infof("Received request to host %v", r.Host)
	if r.URL.Path == "/healthz" {
		log.Info("Received health check")
		w.WriteHeader(http.StatusOK)
		w.Write(healthResponse)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("contentType=%s, expect application/json", contentType)
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write(wrongContentResponse)
		return
	}
	log.Info("Correct ContentType")
	receivedAdmissionReview := v1beta1.AdmissionReview{}
	returnedAdmissionReview := v1beta1.AdmissionReview{}
	var err error
	if err = Decode(body, &receivedAdmissionReview); err != nil {
		log.Error(err)
		admissionResponse = admissionResponseError(err)

	} else if err = validateRequest(receivedAdmissionReview); err != nil {
		log.Error(err)
		admissionResponse = admissionResponseError(err)
	} else {
		log.Info("Test passed, mutating")
		admissionResponse = ah.mutate(receivedAdmissionReview)
		admissionResponse.UID = receivedAdmissionReview.Request.UID
	}
	returnedAdmissionReview.Response = admissionResponse
	responseInBytes, err := json.Marshal(returnedAdmissionReview)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("Writing response")
	if _, err := w.Write(responseInBytes); err != nil {
		log.Error(err)
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

func (ah AppMeshHandler) mutate(receivedAdmissionReview v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	admissionResponse := v1beta1.AdmissionResponse{}
	raw := receivedAdmissionReview.Request.Object.Raw
	pod := corev1.Pod{}
	var ports string
	var name string

	admissionResponse.Allowed = true
	if err := json.Unmarshal(raw, &pod); err != nil {
		log.Error(err)
		return admissionResponseError(err)
	}

	// If sidecar injection is disabled in the annotation, we skip mutating
	switch strings.ToLower(pod.ObjectMeta.Annotations[sidecarInjectAnnotation]) {
	case "disabled":
		log.Info("sidecar inject is disabled. Skipping mutating")
		return &admissionResponse
	}

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
	if v, ok := pod.ObjectMeta.Annotations[virtualNodeNameAnnotation]; ok {
		name = v
	} else {
		// if virtual router name isn't specified in the pod annotation, use the controller owner name instead.
		// https://github.com/awslabs/aws-app-mesh-inject/issues/4
		if controllerName := getControllerNameForPod(pod, receivedAdmissionReview.Request.Namespace); controllerName != nil {
			name = fmt.Sprintf("%s-%s", *controllerName, receivedAdmissionReview.Request.Namespace)
		} else {
			log.Info(ErrNoName)
			return &admissionResponse
		}
	}
	log.Info("injecting appmesh pod")
	log.Infof("Retrieving patch for mesh %v, in region %v, for pod %v, on ports %v, ecr-secrets: %v",
		ah.Config.Name,
		ah.Config.Region,
		name,
		ports,
		ah.Config.EcrSecret,
	)
	admissionResponse.Patch = patch.GetPatch(
		len(pod.Spec.InitContainers),
		len(pod.Spec.Containers),
		len(pod.Spec.ImagePullSecrets),
		ah.Config.Name,
		ah.Config.Region,
		name,
		ports,
		ah.Config.LogLevel,
		ah.Config.EcrSecret,
	)
	log.Infof("Received patch %v", string(admissionResponse.Patch))
	pt := v1beta1.PatchTypeJSONPatch
	admissionResponse.PatchType = &pt
	return &admissionResponse
}

// get the name of the controller that created the pod.
func getControllerNameForPod(pod corev1.Pod, namespace string) *string {
	controllerRef := metav1.GetControllerOf(pod.GetObjectMeta())
	if controllerRef == nil {
		// An orphan
		return nil
	}

	if controllerRef.Kind != apps.SchemeGroupVersion.WithKind("ReplicaSet").Kind {
		// The pod is not owned by a replica set. Return the controller name directly
		return &controllerRef.Name
	}
	rs, err := clientset.AppsV1().ReplicaSets(namespace).Get(controllerRef.Name, metav1.GetOptions{})
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

func NewServer(c config.Config) (*http.Server, error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	server := NewServerNoSSL(c)
	sCert, err := tls.LoadX509KeyPair(tlscert, tlskey)
	if err != nil {
		return server, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
	server.TLSConfig = tlsConfig
	server.Addr = ":8080"
	return server, nil
}

func NewServerNoSSL(c config.Config) *http.Server {
	return &http.Server{
		Handler: AppMeshHandler{
			Config: c,
		},
		Addr: ":8080",
	}
}

func Decode(b []byte, o runtime.Object) error {
	deserializer := codecs.UniversalDeserializer()
	_, _, err := deserializer.Decode(b, nil, o)
	return err
}
