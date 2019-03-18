package config

type Config struct {
	Name      string
	Region    string
	LogLevel  string
	EcrSecret bool
	TlsCert   string
	TlsKey    string
}
