package patch

import "fmt"

const (
	initContainer = `
	    {
	      "name": "proxyinit",
	      "image": "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest",
	      "securityContext": {
		  "capabilities": {
		      "add": [
			"NET_ADMIN"
		      ]
		  }
		},
	      "env": [
		{
		  "name": "APPMESH_START_ENABLED",
		  "value": "1"
		},
		{
		  "name": "APPMESH_IGNORE_UID",
		  "value": "1337"
		},
		{
		  "name": "APPMESH_ENVOY_INGRESS_PORT",
		  "value": "15000"
		},
		{
		  "name": "APPMESH_ENVOY_EGRESS_PORT",
		  "value": "15001"
		},
		{
		  "name": "APPMESH_APP_PORTS",
		  "value": "%v"
		},
		{
		  "name": "APPMESH_EGRESS_IGNORED_IP",
		  "value": "169.254.169.254"
		}
	      ]
	    }
	`
	container = `
	    {
	      "name": "envoy",
	      "image": "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta",
	      "securityContext":
		{
		  "runAsUser": 1337
		},
	      "ports": [
		{
		  "containerPort": 9901,
		  "name": "stats",
		  "protocol": "TCP"
		}
	      ],
	      "env": [
		{
		  "name": "APPMESH_VIRTUAL_NODE_NAME",
		  "value": "mesh/%v/virtualNode/%v"
		},
		{
		  "name": "ENVOY_LOG_LEVEL",
		  "value": "%v"
		},
		{
		  "name": "AWS_REGION",
		  "value": "%v"
		}
	      ],
	      "resources": {
		  	"requests": {
		  	"cpu": "10m",
		  	"memory": "32Mi"
		  	}
	      }
	    }
	`
	ecr_secret = `{"name": "appmesh-ecr-secret"}`
	create     = `{"op":"add","path":"/spec/%v", "value": [%v]}`
	add        = `{"op":"add","path":"/spec/%v/-", "value": %v}`
)

func GetPatch(i, c, e int, mesh, region, virtualnode, ports, log string, ecr bool) []byte {
	//TODO: cleanup all this code
	init := fmt.Sprintf(initContainer, ports)
	cont := fmt.Sprintf(container, mesh, virtualnode, log, region)
	var initPatch string
	var contPatch string
	if i > 0 {
		initPatch = fmt.Sprintf(add, "initContainers", init)
	} else {
		initPatch = fmt.Sprintf(create, "initContainers", init)
	}
	if c > 0 {
		contPatch = fmt.Sprintf(add, "containers", cont)
	} else {
		contPatch = fmt.Sprintf(create, "containers", cont)
	}
	if ecr {
		var ecrPatch string
		if e > 0 {
			ecrPatch = fmt.Sprintf(add, "imagePullSecrets", ecr_secret)
		} else {
			ecrPatch = fmt.Sprintf(create, "imagePullSecrets", ecr_secret)
		}
		patch := fmt.Sprintf("[%v,%v,%v]", initPatch, contPatch, ecrPatch)
		return []byte(patch)
	}
	patch := fmt.Sprintf("[%v,%v]", initPatch, contPatch)
	return []byte(patch)
}
