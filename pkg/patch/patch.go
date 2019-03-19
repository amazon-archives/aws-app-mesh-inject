package patch

import "fmt"

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

	sidecarPatch, err := renderSidecar(meta.Sidecar)
	if err != nil {
		return []byte(patch), err
	}

	if meta.AppendSidecar {
		sidecarPatch = fmt.Sprintf(add, "containers", sidecarPatch)
	} else {
		sidecarPatch = fmt.Sprintf(create, "containers", sidecarPatch)
	}

	if meta.HasImagePullSecret {
		var ecrPatch string
		if meta.AppendImagePullSecret {
			ecrPatch = fmt.Sprintf(add, "imagePullSecrets", ecrSecret)
		} else {
			ecrPatch = fmt.Sprintf(create, "imagePullSecrets", ecrSecret)
		}
		patch = fmt.Sprintf("[%v,%v,%v]", initPatch, sidecarPatch, ecrPatch)
	} else {
		patch = fmt.Sprintf("[%v,%v]", initPatch, sidecarPatch)
	}

	return []byte(patch), nil
}
