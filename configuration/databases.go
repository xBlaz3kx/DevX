package configuration

import "github.com/xBlaz3kx/DevX/tls"

type (
	// Redis database configuration with TLS
	Redis struct {
		Address  string  `yaml:"address" json:"address" mapstructure:"address"`
		Password string  `yaml:"password" json:"password" mapstructure:"password"`
		TLS      tls.TLS `mapstructure:"tls" yaml:"tls" json:"tls"`
	}

	// Primary Database configuration
	Database struct {
		Type       string  `json:"type" yaml:"type" mapstructure:"type" required:""`
		URI        string  `json:"uri,omitempty" yaml:"uri" mapstructure:"uri"`
		Host       string  `json:"host,omitempty" yaml:"host" mapstructure:"host"`
		Username   string  `json:"username,omitempty" yaml:"username" mapstructure:"username"`
		Password   string  `json:"password,omitempty" yaml:"password" mapstructure:"password"`
		Port       int     `json:"port,omitempty" yaml:"port" mapstructure:"port"`
		Database   string  `json:"database,omitempty" yaml:"database" mapstructure:"database"`
		ReplicaSet string  `json:"replicaSet,omitempty" yaml:"replicaSet" mapstructure:"replicaSet"`
		TLS        tls.TLS `json:"tls" yaml:"tls" mapstructure:"tls"`
	}

	// InfluxDB database configuration with TLS
	InfluxDB struct {
		URL          string  `yaml:"url" json:"url" mapstructure:"url"`
		Organization string  `yaml:"organization" json:"organization" mapstructure:"organization"`
		Bucket       string  `yaml:"bucket" json:"bucket" mapstructure:"bucket"`
		AccessToken  string  `yaml:"accessToken" json:"accessToken" mapstructure:"accessToken"`
		TLS          tls.TLS `mapstructure:"tls" yaml:"tls" json:"tls"`
	}
)
