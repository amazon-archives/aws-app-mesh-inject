package patch

import (
	"bufio"
	"bytes"
	"text/template"
)

const secretVolumeMount = `
{
  "mountPath": "{{ .MountPath }}",
  "name": "{{ .SecretName }}",
  "readOnly": true
}
`

const secretVolume = `
{
  "name": "{{ .SecretName }}",
  "secret":
  {
    "secretName": "{{ .SecretName }}"
  }
}
`

func renderSecretVolumeMount(secretMount SecretMount) (string, error) {
	return renderSecretCommon(secretMount, secretVolumeMount)
}


func renderSecretVolume(secretMount SecretMount) (string, error) {
	return renderSecretCommon(secretMount, secretVolume)
}

func renderSecretCommon(secretMount SecretMount, tpl string) (string, error) {
	tmpl, err := template.New("mount").Parse(tpl)
	if err != nil {
		return "", err
	}

	var data bytes.Buffer
	writer := bufio.NewWriter(&data)
	if err := tmpl.Execute(writer, secretMount); err != nil {
		return "", err
	}
	err = writer.Flush()
	if err != nil {
		return "", err
	}

	return data.String(), nil
}
