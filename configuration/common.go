package configuration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

// Default values and dirs
const (
	CurrentDir    = "."
	ConfigDir     = "./config"
	CmdDir        = "../../config/"
	ConfigName    = "config"
	FileExtension = "yaml"
)

const (
	TracingFlag = "tracing"
	MetricsFlag = "metrics"
)

// GetConfiguration loads the configuration from Viper and validates it.
func GetConfiguration(viper *viper.Viper, configStruct interface{}) {
	// Load configuration from file
	err := viper.Unmarshal(configStruct)
	if err != nil {
		zap.L().Fatal("Cannot unmarshall", zap.Error(err))
	}

	// Validate configuration
	validationErr := validator.New().Struct(configStruct)
	if validationErr != nil {

		var errs validator.ValidationErrors
		errors.As(validationErr, &errs)

		hasErr := false
		for _, fieldError := range errs {
			zap.L().Error("Validation failed on field", zap.Error(fieldError))
			hasErr = true
		}

		if hasErr {
			zap.L().Fatal("Validation of the config failed")
		}
	}
}

// SetupEnv sets up the environment variables for the service.
func SetupEnv(serviceName string) {
	viper.SetEnvPrefix(serviceName)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// SetDefaults sets the default values for the app configuration.
func SetDefaults(serviceName string) {
	viper.SetDefault(MetricsEnable, false)
	viper.SetDefault(MetricsEndpoint, "/metrics")
	viper.SetDefault(TracingEnable, false)
}

// InitConfig initializes the configuration for the service, either from file or from ETCD (if ETCD was set).
func InitConfig(configurationFilePath string, additionalDirs ...string) {
	// Set defaults for searching for config files
	viper.SetConfigName(ConfigName)
	viper.SetConfigType(FileExtension)
	viper.AddConfigPath(ConfigDir)
	viper.AddConfigPath(CmdDir)
	viper.AddConfigPath(CurrentDir)

	for _, dir := range additionalDirs {
		viper.AddConfigPath(dir)
	}

	// Check if path is specified
	if stringUtils.IsNotEmpty(configurationFilePath) {
		viper.SetConfigFile(configurationFilePath)
	}

	// Check if ETCD is configured
	etcdAddress := viper.GetString(EtcdAddress)
	if stringUtils.IsNotEmpty(etcdAddress) {
		zap.L().Info("Using ETCD for configuration")

		prefix := viper.GetString(EtcdPrefix)
		err := viper.AddRemoteProvider("etcd3", etcdAddress, fmt.Sprintf("/%s/%s.%s", prefix, ConfigName, FileExtension))
		if err != nil {
			zap.L().With(zap.Error(err)).Error("Cannot configure remote provider")
		}

		err = viper.ReadRemoteConfig()
		if err != nil {
			zap.L().With(zap.Error(err)).Error("Cannot read from remote provider")
		}
	}

	// Read the configuration
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			zap.L().With(zap.Error(err)).Error("Config file not found")
		} else {
			zap.L().With(zap.Error(err)).Fatal("Something went wrong")
		}
	}
}
