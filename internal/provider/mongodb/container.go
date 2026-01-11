package mongodb

import (
	"context"
	"fmt"
	"sync"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Container manages a MongoDB testcontainer with replica set support
type Container struct {
	container *mongodb.MongoDBContainer
	client    *mongo.Client
	connStr   string
	mu        sync.Mutex
}

// NewContainer creates a new MongoDB container manager
func NewContainer() *Container {
	return &Container{}
}

// Start launches the MongoDB container with replica set support
func (c *Container) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.container != nil {
		return nil // Already running
	}

	// Start MongoDB with replica set for transaction support
	container, err := mongodb.Run(ctx,
		"mongo:7.0",
		mongodb.WithReplicaSet("rs0"),
	)
	if err != nil {
		return fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	c.container = container

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		c.Stop(ctx)
		return fmt.Errorf("failed to get connection string: %w", err)
	}
	c.connStr = connStr

	// Create MongoDB client
	clientOpts := options.Client().ApplyURI(connStr)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		c.Stop(ctx)
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		c.Stop(ctx)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.client = client
	return nil
}

// Stop terminates the MongoDB container
func (c *Container) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		if err := c.client.Disconnect(ctx); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to disconnect client: %v\n", err)
		}
		c.client = nil
	}

	if c.container != nil {
		if err := c.container.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
		c.container = nil
	}

	c.connStr = ""
	return nil
}

// IsRunning returns whether the container is running
func (c *Container) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container != nil && c.client != nil
}

// Client returns the MongoDB client
func (c *Container) Client() *mongo.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client
}

// ConnectionString returns the connection string
func (c *Container) ConnectionString() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connStr
}

// Database returns a database handle
func (c *Container) Database(name string) *mongo.Database {
	if c.client == nil {
		return nil
	}
	return c.client.Database(name)
}
