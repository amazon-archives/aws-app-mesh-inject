package patch

import (
	"fmt"
	"strings"
)

const (
	ecrSecret = `{"name": "appmesh-ecr-secret"}`
	create    = `{"op":"add","path":"/spec/%v", "value": [%v]}`
	add       = `{"op":"add","path":"/spec/%v/-", "value": %v}`
)

type Meta struct {
	AppendInit            bool
	AppendSidecar         bool
	AppendImagePullSecret bool
	HasImagePullSecret    bool
	Init                  InitMeta
	Sidecar               SidecarMeta
}

func GeneratePatch(meta Meta) ([]byte, error) {
	var patch string

	initPatch, err := renderInit(meta.Init)
	if err != nil {
		return []byte(patch), err
	}

	if meta.AppendInit {
		initPatch = fmt.Sprintf(add, "initContainers", initPatch)
	} else {
		initPatch = fmt.Sprintf(create, "initContainers", initPatch)
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

	var patches []string
	patches = append(patches, initPatch)
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
