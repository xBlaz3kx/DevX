package configuration

// Cross Origin Resource Sharing configuration
type CORS struct {
	// IsEnabled is the flag to enable/disable CORS
	IsEnabled bool `yaml:"enabled" json:"enabled,omitempty" mapstructure:"enabled"`

	// AllowedOrigins is the list of allowed origins
	AllowedOrigins []string `yaml:"allowedOrigins" json:"allowedOrigins,omitempty" mapstructure:"allowedOrigins"`
}
