package tests

import (
	"context"

	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	devxCfg "github.com/xBlaz3kx/DevX/configuration"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DatabaseName = "alpr-cloud"

type MongoContainer struct {
	client *mongo.Client
	*mongodb.MongoDBContainer
}

func NewMongoContainer(ctx context.Context) (*MongoContainer, error) {
	mongodbContainer, err := mongodb.Run(ctx, "mongo:8", mongodb.WithReplicaSet("rs0"))
	if err != nil {
		return nil, err
	}

	return &MongoContainer{MongoDBContainer: mongodbContainer}, nil
}

func (c *MongoContainer) CreateClient(ctx context.Context) (*mongo.Client, *checks.MongoCheck, func(), error) {
	// Get the connection string
	connectionString, err := c.MongoDBContainer.ConnectionString(ctx)
	if err != nil {
		return nil, nil, func() {}, err
	}

	cfg := devxCfg.Database{URI: connectionString, Database: DatabaseName}

	clientOpts := options.Client()
	clientOpts.ApplyURI(cfg.URI)
	clientOpts.SetAppName("testing-client")

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, func() {}, err
	}

	checker := checks.NewMongoCheck(10, client)

	return client, checker, func() {
		_ = client.Disconnect(ctx)
	}, nil
}

func (c *MongoContainer) Client() *mongo.Client {
	return c.client
}