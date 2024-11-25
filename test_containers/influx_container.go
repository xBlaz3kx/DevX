package tests

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	InfluxUsername = "admin"
	InfluxPassword = "examplePassword"
	InfluxOrg      = "dev"
	InfluxBucket   = "integration"
	InfluxToken    = "adminToken"
)

type InfluxContainer struct {
	testcontainers.Container
	URI string
}

// Creates and runs an InfluxDB V2 container with the following environment variables:
// "DOCKER_INFLUXDB_INIT_USERNAME":     "admin",
// "DOCKER_INFLUXDB_INIT_PASSWORD":     "examplePassword",
// "DOCKER_INFLUXDB_INIT_ORG":          "dev",
// "DOCKER_INFLUXDB_INIT_BUCKET":       "integration",
// "DOCKER_INFLUXDB_INIT_ADMIN_TOKEN":  "adminToken",
func NewInfluxContainer(ctx context.Context, networkName string) (*InfluxContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         "influxdb",
		Image:        "influxdb:2.7.8",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor:   wait.ForHTTP("/health"),
		Networks:     []string{"host"},
		Env: map[string]string{
			"DOCKER_INFLUXDB_INIT_MODE":         "setup",
			"DOCKER_INFLUXDB_INIT_USERNAME":     InfluxUsername,
			"DOCKER_INFLUXDB_INIT_PASSWORD":     InfluxPassword,
			"DOCKER_INFLUXDB_INIT_ORG":          InfluxOrg,
			"DOCKER_INFLUXDB_INIT_BUCKET":       InfluxBucket,
			"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN":  InfluxToken,
			"DOCKER_INFLUXDB_INIT_AUTH_ENABLED": "true",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &InfluxContainer{Container: container}, nil
}

func (i *InfluxContainer) GetConnectionString(ctx context.Context) (string, error) {
	if i.URI != "" {
		return i.URI, nil
	}

	ip, err := i.Container.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := i.Container.MappedPort(ctx, "8086")
	if err != nil {
		return "", err
	}

	i.URI = fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())
	return i.URI, nil
}

func (i *InfluxContainer) GetInfluxClient(ctx context.Context) (influxdb2.Client, error) {
	connectionString, err := i.GetConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return influxdb2.NewClient(connectionString, InfluxToken), nil
}
