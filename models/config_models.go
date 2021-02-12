package models

// Config ...
type Config struct {
	HTTP struct {
		Port       int    `yaml:"port"`
		IsTLS      bool   `yaml:"is_tls"`
		ServerCert string `yaml:"server_crt"`
		ServerKey  string `yaml:"server_key"`
	} `yaml:"http"`
}
