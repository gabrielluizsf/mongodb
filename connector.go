package mongodb

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connector defines a generic interface for establishing a connection
// to a resource of type T.
type Connector[T any] interface {
	// Connect initializes the connection and returns the connected resource.
	Connect() (*T, error)
}

// DatabaseConnector implements Connector for MongoDB databases.
// It holds all configuration required to create a MongoDB client
// and access a specific database.
type DatabaseConnector struct {
	// DatabaseName is the name of the MongoDB database to connect to.
	DatabaseName string

	// URI is the MongoDB connection string.
	URI string

	// Options holds optional database-level configuration.
	Options *options.DatabaseOptions

	// ClientOptions holds optional client-level configuration.
	// If provided, it overrides the default options created from the URI.
	ClientOptions *options.ClientOptions

	// Client holds the underlying MongoDB client instance created during Connect.
	// It can be used to access client-level operations or to close the connection
	// when it is no longer needed.
	Client *mongo.Client
}

// NewConnector creates a new MongoDB database connector using
// the provided database name and connection URI.
func NewConnector(
	databaseName string,
	uri string,
) Connector[mongo.Database] {
	c := &DatabaseConnector{
		DatabaseName: databaseName,
		URI:          uri,
	}
	return c
}

// Connect creates a MongoDB client, applies the configured options,
// and returns a handle to the configured database.
func (c *DatabaseConnector) Connect() (*mongo.Database, error) {
	opts := options.Client().ApplyURI(c.URI)
	if c.ClientOptions != nil {
		opts = c.ClientOptions
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	c.Client = client
	if c.Options != nil {
		return client.Database(c.DatabaseName, BuildDatabaseOptions(c.Options)), nil
	}

	return client.Database(c.DatabaseName), nil
}
