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

	return []byte(fmt.Sprintf("[%s]", strings.Join(patches, ","))), nil
}
