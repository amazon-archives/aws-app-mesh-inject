package patch

import (
	"fmt"
	"strings"

	"github.com/aws/aws-app-mesh-inject/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ecrSecret        = `{"name": "appmesh-ecr-secret"}`
	create           = `{"op":"add","path":"/spec/%v", "value": [%v]}`
	createElement    = `{"op":"add","path":"/spec/%v", "value": %v}`
	add              = `{"op":"add","path":"/spec/%v/-", "value": %v}`
	createAnnotation = `{"op":"add","path":"/metadata/annotations","value":{"%s": "%s"}}`
	updateAnnotation = `{"op":"%s","path":"/metadata/annotations/%s","value":"%s"}`

	defaultFSGroup int64 = 1337
)

type Meta struct {
	AppendInit            bool
	AppendSidecar         bool
	AppendImagePullSecret bool
	HasImagePullSecret    bool
	InjectFSGroup         bool
	Init                  InitMeta
	Sidecar               SidecarMeta
	PodMetadata           metav1.ObjectMeta
}

func GeneratePatch(meta Meta) ([]byte, error) {
	var patches []string
	var emptyPatch []byte

	if isAppMeshCNIEnabled(meta) {
		patches = append(patches, appMeshCNIAnnotationsPatch(meta)...)
	} else {
		initPatch, err := renderInit(meta.Init)
		if err != nil {
			return emptyPatch, err
		}

		if meta.AppendInit {
			initPatch = fmt.Sprintf(add, "initContainers", initPatch)
		} else {
			initPatch = fmt.Sprintf(create, "initContainers", initPatch)
		}
		patches = append(patches, initPatch)
	}

	if meta.InjectFSGroup {
		patches = append(patches, podFSGroupPatch(defaultFSGroup))
	}

	var sidecarPatches []string
	sidecars, err := renderSidecars(meta.Sidecar)
	if err != nil {
		return emptyPatch, err
	}

	if meta.AppendSidecar {
		//will generate values of the form [{"op":"add","path":"/spec/containers/-", {...}},...]
		for i := range sidecars {
			sidecarPatches = append(sidecarPatches, fmt.Sprintf(add, "containers", sidecars[i]))
		}
	} else {
		//will generate values of the form {"op":"add","path":"/spec/containers", [{...},{...}]}
		sidecarPatches = append(sidecarPatches, fmt.Sprintf(create, "containers", strings.Join(sidecars, ",")))
	}

	patches = append(patches, sidecarPatches...)
	if meta.HasImagePullSecret {
		var ecrPatch string
		if meta.AppendImagePullSecret {
			ecrPatch = fmt.Sprintf(add, "imagePullSecrets", ecrSecret)
		} else {
			ecrPatch = fmt.Sprintf(create, "imagePullSecrets", ecrSecret)
		}
		patches = append(patches, ecrPatch)
	}

	if meta.Sidecar.EnableDatadogTracing {
		// add an empty dir volume for the Envoy static config
		volumePatch := fmt.Sprintf(add, "volumes", renderDatadogConfigVolume())
		patches = append(patches, volumePatch)

		// add an init container that writes the Envoy static config to the empty dir volume
		datadogInit, err := renderDatadogInitContainer(meta.Sidecar.DatadogAddress, meta.Sidecar.DatadogPort)
		if err != nil {
			return emptyPatch, err
		}

		j := fmt.Sprintf(add, "initContainers", datadogInit)
		patches = append(patches, j)
	}

	if meta.Sidecar.EnableJaegerTracing {
		// add an empty dir volume for the Envoy static config
		volumePatch := fmt.Sprintf(add, "volumes", renderJaegerConfigVolume())
		patches = append(patches, volumePatch)

		// add an init container that writes the Envoy static config to the empty dir volume
		jaegerInit, err := renderJaegerInitContainer(meta.Sidecar.JaegerAddress, meta.Sidecar.JaegerPort)
		if err != nil {
			return emptyPatch, err
		}

		j := fmt.Sprintf(add, "initContainers", jaegerInit)
		patches = append(patches, j)
	}

	fmt.Println(patches)

	return []byte(fmt.Sprintf("[%s]", strings.Join(patches, ","))), nil
}

func appMeshCNIAnnotationsPatch(meta Meta) []string {
	newAnnotations := map[string]string{
		config.AppMeshEgressIgnoredIPsAnnotation:   meta.Init.IgnoredIPs,
		config.AppMeshEgressIgnoredPortsAnnotation: meta.Init.EgressIgnoredPorts,
		config.AppMeshPortsAnnotation:              meta.Init.Ports,
		config.AppMeshSidecarInjectAnnotation:      "enabled",
		//Below settings are fixed as per current App Mesh runtime behavior
		config.AppMeshIgnoredUIDAnnotation:       config.AppMeshProxyUID,
		config.AppMeshProxyEgressPortAnnotation:  config.AppMeshProxyEgressPort,
		config.AppMeshProxyIngressPortAnnotation: config.AppMeshProxyIngressPort,
	}
	return annotationsPatches(meta.PodMetadata.Annotations, newAnnotations)
}

func podFSGroupPatch(fsGroup int64) string {
	patch := fmt.Sprintf(createElement, "securityContext/fsGroup", fsGroup)
	return patch
}

func annotationsPatches(existingAnnotations map[string]string, newAnnotations map[string]string) (patches []string) {
	for key, value := range newAnnotations {
		if existingAnnotations == nil {
			//first one will be create, subsequent will be updates
			existingAnnotations = map[string]string{}
			patches = append(patches, fmt.Sprintf(createAnnotation, key, value))
		} else {
			op := "add"
			if existingAnnotations[key] != "" {
				op = "replace"
			}
			patches = append(patches, fmt.Sprintf(updateAnnotation, op, escapeJSONPointer(key), value))
		}
	}

	return patches
}

func escapeJSONPointer(key string) string {
	s0 := strings.ReplaceAll(key, "~", "~0")
	return strings.ReplaceAll(s0, "/", "~1")
}

func isAppMeshCNIEnabled(meta Meta) bool {
	if v, ok := meta.PodMetadata.Annotations[config.AppMeshCNIAnnotation]; ok {
		return v == "enabled"
	}
	//Fargate platform has appmesh-cni enabled by default
	if v, ok := meta.PodMetadata.Labels[config.FargateProfileLabel]; ok {
		return len(v) > 0
	}
	return false
}
