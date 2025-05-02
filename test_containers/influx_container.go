package tests

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
)

const (
	InfluxUsername = "admin"
	InfluxPassword = "examplePassword"
	InfluxOrg      = "dev"
	InfluxBucket   = "integration"
	InfluxToken    = "adminToken"
)

type InfluxDbV2Container struct {
	*influxdb.InfluxDbContainer
}

// Creates and runs an InfluxDB V2 container with the following environment variables:
// "DOCKER_INFLUXDB_INIT_USERNAME":     "admin",
// "DOCKER_INFLUXDB_INIT_PASSWORD":     "examplePassword",
// "DOCKER_INFLUXDB_INIT_ORG":          "dev",
// "DOCKER_INFLUXDB_INIT_BUCKET":       "integration",
// "DOCKER_INFLUXDB_INIT_ADMIN_TOKEN":  "adminToken",
func NewInfluxV2Container(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InfluxDbV2Container, error) {
	influxdbContainer, err := influxdb.Run(
		ctx,
		"influxdb:2.7.11",
		influxdb.WithV2Auth(InfluxOrg, InfluxBucket, InfluxUsername, InfluxPassword),
		influxdb.WithV2AdminToken(InfluxToken),
	)
	if err != nil {
		return nil, err
	}

	return &InfluxDbV2Container{InfluxDbContainer: influxdbContainer}, nil
}

func (i *InfluxDbV2Container) GetInfluxClient(ctx context.Context) (influxdb2.Client, error) {
	connectionString, err := i.InfluxDbContainer.ConnectionUrl(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get connection URL")
	}

	return influxdb2.NewClient(connectionString, InfluxToken), nil
}
