/*
  Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.

  Licensed under the Apache License, Version 2.0 (the "License").
  You may not use this file except in compliance with the License.
  A copy of the License is located at

      http://www.apache.org/licenses/LICENSE-2.0

  or in the "license" file accompanying this file. This file is distributed
  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
  express or implied. See the License for the specific language governing
  permissions and limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"time"

	"github.com/aws/aws-app-mesh-inject/pkg/config"
	"github.com/aws/aws-app-mesh-inject/pkg/signals"
	"github.com/aws/aws-app-mesh-inject/pkg/webhook"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	masterURL  string
	kubeconfig string
	enableTLS  bool
	cfg        config.Config
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&cfg.MeshName, "mesh-name", os.Getenv("APPMESH_NAME"), "AWS App Mesh name")
	flag.StringVar(&cfg.Region, "region", os.Getenv("APPMESH_REGION"), "AWS App Mesh region")
	flag.StringVar(&cfg.LogLevel, "log-level", os.Getenv("APPMESH_LOG_LEVEL"), "AWS App Mesh envoy log level")
	flag.BoolVar(&cfg.EcrSecret, "ecr-secret", false, "Inject AWS app mesh pull secrets")
	flag.IntVar(&cfg.Port, "port", 8080, "Webhook port")
	flag.StringVar(&cfg.TlsCert, "tlscert", "/etc/webhook/certs/cert.pem", "Location of TLS Cert file.")
	flag.StringVar(&cfg.TlsKey, "tlskey", "/etc/webhook/certs/key.pem", "Location of TLS key file.")
	flag.BoolVar(&enableTLS, "enable-tls", true, "Enable TLS.")
	flag.StringVar(&cfg.SidecarImage, "sidecar-image", "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.9.1.0-prod", "Envoy sidecar container image.")
	flag.StringVar(&cfg.SidecarCpu, "sidecar-cpu-requests", "10m", "Envoy sidecar CPU resources requests.")
	flag.StringVar(&cfg.SidecarMemory, "sidecar-memory-requests", "32Mi", "Envoy sidecar memory resources requests.")
	flag.StringVar(&cfg.InitImage, "init-image", "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest", "Init container image.")
	flag.StringVar(&cfg.IgnoredIPs, "ignored-ips", "169.254.169.254", "Init container ignored IPs.")
	flag.BoolVar(&cfg.InjectXraySidecar, "inject-xray-sidecar", false, "Enable Envoy X-Ray tracing integration and injects xray-daemon as sidecar")
	flag.BoolVar(&cfg.EnableStatsTags, "enable-stats-tags", false, "Enable Envoy to tag stats")
}

func main() {
	flag.Set("logtostderr", "true")
	klog.InitFlags(nil)
	flag.Parse()

	// init Kubernetes config
	kubeConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}

	// init Kubernetes client
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	// init Kubernetes deserializer
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	admissionregistrationv1beta1.AddToScheme(scheme)
	codecs := serializer.NewCodecFactory(scheme)
	kubeDecoder := codecs.UniversalDeserializer()

	// set default region
	if cfg.Region == "" {
		// Use region from ec2 metadata service by default
		s, err := session.NewSession(&aws.Config{})
		if err != nil {
			klog.Fatal("Failed to create an aws config session", err)
		}
		metadata := ec2metadata.New(s)
		cfg.Region, err = metadata.Region()
		if err != nil {
			klog.Fatal("Failed to determine the region from ec2 metadata", err)
		}
	}

	// init webhook HTTP server
	srv := &webhook.Server{
		Config:      cfg,
		KubeClient:  kubeClient,
		KubeDecoder: kubeDecoder,
	}

	// start HTTP server
	stopCh := signals.SetupSignalHandler()
	go srv.ListenAndServe(enableTLS, 5*time.Second, stopCh)

	// wait for SIGTERM
	<-stopCh
}
