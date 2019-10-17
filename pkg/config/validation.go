package config

// MultipleTracer checks if more than one tracer is configured.
func MultipleTracer(config Config) bool {
	j := config.EnableJaegerTracing
	d := config.EnableDatadogTracing
	x := config.InjectXraySidecar

	return ((j && d) || (d && x) || (j && x))
}
