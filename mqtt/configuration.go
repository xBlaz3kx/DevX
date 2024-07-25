package mqtt

import (
	"github.com/xBlaz3kx/DevX/configuration"
)

// Configuration is the configuration for the MQTT client
type Configuration struct {
	Address  string            `validate:"required" json:"address" yaml:"address"`
	Username string            `fig:"username" yaml:"username"`
	Password string            `fig:"password" yaml:"password"`
	ClientId string            `fig:"clientId" validate:"required" yaml:"clientId"`
	TLS      configuration.TLS `fig:"tls" validate:"required" yaml:"tls"`
}
