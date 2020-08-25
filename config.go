package main

// Config manages the service configuration
type Config struct {
	Name  string   `envconfig:"GUZZLER_NAME" required:"true"`
	Bind  string   `envconfig:"GUZZLER_BIND" default:"127.0.0.1:8080" required:"false"`
	Peers []string `envconfig:"GUZZLER_PEERS" required:"true" split_words:"true"` /* ","-separated list of peers */
}
