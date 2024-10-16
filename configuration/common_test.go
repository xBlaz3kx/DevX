package configuration

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xBlaz3kx/DevX/tls"
)

type exampleConfiguration struct {
	Example  string   `json:"example" yaml:"example"`
	InfluxDB InfluxDB `json:"influxdb" yaml:"influxdb"`
}

func TestGetConfiguration(t *testing.T) {
	//todo fix
	example := exampleConfiguration{
		Example: "test",
		InfluxDB: InfluxDB{
			URL:          "http://localhost:8086",
			Organization: "example",
			Bucket:       "example",
			AccessToken:  "accessToken",
			TLS: tls.TLS{
				IsEnabled: false,
			},
		},
	}
	marshal, err := json.Marshal(example)
	assert.NoError(t, err)

	// Inject the configuration
	viper.SetConfigType("json")
	err = viper.ReadConfig(bytes.NewBuffer(marshal))
	assert.NoError(t, err)

	getExampleConfig := exampleConfiguration{}
	GetConfiguration(viper.GetViper(), &getExampleConfig)

	assert.Equal(t, example, getExampleConfig)
}

func TestGetConfiguration_NotLoaded(t *testing.T) {
	getExampleConfig := exampleConfiguration{}
	GetConfiguration(viper.New(), &getExampleConfig)

	assert.Empty(t, getExampleConfig)
}

func TestSetupEnv(t *testing.T) {
	SetupEnv("test")

	err := os.Setenv("TEST_EXAMPLE", "test")
	assert.NoError(t, err)

	val := viper.GetString("EXAMPLE")
	assert.Equal(t, "test", val)

	// Ok case
	err = os.Setenv("EXAMPLE2", "test")
	assert.NoError(t, err)

	// Env with no prefix - should not be found
	val = viper.GetString("EXAMPLE2")
	assert.Empty(t, val)

	// Nested env
	err = os.Setenv("TEST_EXAMPLE123_EXAMPLE", "123EXAMPLETEST")
	assert.NoError(t, err)

	val = viper.GetString("EXAMPLE123.EXAMPLE")
	assert.Equal(t, "123EXAMPLETEST", val)
}

func TestSetDefaults(t *testing.T) {
	SetDefaults("test")

	val := viper.GetBool(MetricsEnable)
	assert.False(t, val)

	val = viper.GetBool(TracingEnable)
	assert.False(t, val)
}

func TestInitConfig(t *testing.T) {
	testCfg := exampleConfiguration{
		Example: "test",
		InfluxDB: InfluxDB{
			URL:          "http://localhost:8086",
			Organization: "example",
			Bucket:       "example",
			AccessToken:  "accessToken",
		},
	}
	err := writeToFile("./test.yaml", testCfg)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		InitConfig("./test.yaml")
	})

	cfg := exampleConfiguration{}
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err)
	assert.EqualValues(t, testCfg, cfg)

	err = os.Remove("./test.yaml")
	require.NoError(t, err)

	// Test default values
	dir, err := os.MkdirTemp(".", "test")
	require.NoError(t, err)

	testFileName := filepath.Join(dir, "test.yaml")
	err = writeToFile(testFileName, testCfg)
	defer os.RemoveAll(dir)

	require.NoError(t, err)

	assert.NotPanics(t, func() {
		InitConfig("", "./test123", "/etc/tmp/example")
	})

	// todo test case where file doesn't exist and there are no defaults

	// todo Test with ETCD
	err = os.Setenv("ETCD_ADDRESS", "http://localhost:2379")
	assert.NoError(t, err)
}

func writeToFile(file string, structure any) error {
	// Marshal the struct to JSON
	data, err := json.Marshal(structure)
	if err != nil {
		return err
	}

	// Write the JSON data to a file
	return os.WriteFile(file, data, 0644)
}
