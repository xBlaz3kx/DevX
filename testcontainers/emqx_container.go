package tests

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/xBlaz3kx/DevX/mqtt"
	"github.com/xBlaz3kx/DevX/observability"
)

type EmqxContainer struct {
	testcontainers.Container
	URI string
}

func NewEmqxContainer(ctx context.Context, networkName string) (*EmqxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "emqx/emqx:5.8.0",
		ExposedPorts: []string{"1883/tcp", "8083/tcp"},
		WaitingFor:   wait.ForExec([]string{"/opt/emqx/bin/emqx", "ctl", "status"}),
		Networks:     []string{"host"},
		Env:          map[string]string{},
		// todo provision container with users, permissions etc
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &EmqxContainer{Container: container}, nil
}

func (e *EmqxContainer) GetURI(ctx context.Context) (string, error) {
	if e.URI != "" {
		return e.URI, nil
	}

	ip, err := e.Container.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := e.Container.MappedPort(ctx, "8086")
	if err != nil {
		return "nil", err
	}

	e.URI = fmt.Sprintf("mqtt://%s:%s", ip, mappedPort.Port())
	return e.URI, nil
}

func (e *EmqxContainer) GetMqttClient(ctx context.Context) (mqtt.Client, error) {
	uri, err := e.GetURI(ctx)
	if err != nil {
		return nil, err
	}

	config := mqtt.Configuration{
		Address: uri,
	}

	mqttClient, err := mqtt.NewV3Client(config, observability.NewNoopObservability())
	if err != nil {
		return nil, err
	}

	return mqttClient, nil
}
