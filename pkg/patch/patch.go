package patch

import (
	"fmt"
	"strings"
)

const (
	ecrSecret            = `{"name": "appmesh-ecr-secret"}`
	create               = `{"op":"add","path":"/spec/%v", "value": [%v]}`
	add                  = `{"op":"add","path":"/spec/%v/-", "value": %v}`
	createAnnotation     = `{"op":"add","path":"/metadata/annotations","value":{"%s": "%s"}}`
	updateAnnotation     = `{"op":"%s","path":"/metadata/annotations/%s","value":"%s"}`
	appmeshCNIAnnotation = "appmesh.k8s.aws/appmeshCNI"
)

type Meta struct {
	AppendInit            bool
	AppendSidecar         bool
	AppendImagePullSecret bool
	HasImagePullSecret    bool
	Init                  InitMeta
	Sidecar               SidecarMeta
	PodAnnotations        map[string]string
}

func GeneratePatch(meta Meta) ([]byte, error) {
	var patches []string

	var patch string

	var appmeshCNIEnabled bool
	if v, ok := meta.PodAnnotations[appmeshCNIAnnotation]; ok {
		appmeshCNIEnabled = (v == "enabled")
	}
	if appmeshCNIEnabled {
		patches = append(patches, appMeshCNIAnnotationsPatch(meta)...)
	} else {
		initPatch, err := renderInit(meta.Init)
		if err != nil {
			return []byte(patch), err
		}

		if meta.AppendInit {
			initPatch = fmt.Sprintf(add, "initContainers", initPatch)
		} else {
			initPatch = fmt.Sprintf(create, "initContainers", initPatch)
		}
		patches = append(patches, initPatch)
	}

	var sidecarPatches []string
	sidecars, err := renderSidecars(meta.Sidecar)
	if err != nil {
		return []byte(patch), err
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
			return []byte(patch), err
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
			return []byte(patch), err
		}

		j := fmt.Sprintf(add, "initContainers", jaegerInit)
		patches = append(patches, j)
	}

	fmt.Println(patches)

	return []byte(fmt.Sprintf("[%s]", strings.Join(patches, ","))), nil
}

func appMeshCNIAnnotationsPatch(meta Meta) []string {
	newAnnotations := map[string]string{
		"appmesh.k8s.aws/egressIgnoredIPs":   meta.Init.IgnoredIPs,
		"appmesh.k8s.aws/egressIgnoredPorts": meta.Init.EgressIgnoredPorts,
		"appmesh.k8s.aws/ports":              meta.Init.Ports,
	}
	return annotationsPatches(meta.PodAnnotations, newAnnotations)
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
