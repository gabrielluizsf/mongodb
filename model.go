// Package mongodb provides a generic MongoDB model implementation
package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoModel is a concrete MongoDB-backed implementation of Model.
// T represents the document type stored in the collection.
// C represents a custom result type, mainly used for aggregations.
//
// Using generics avoids reflection, improving performance and type safety.
type mongoModel[T, C any] struct {
	// Name is the MongoDB collection name.
	Name string

	// collection is the underlying MongoDB collection instance.
	collection *mongo.Collection
}

// mongodb is a type alias that binds mongoModel to the generic Model interface.
//
// This improves readability by exposing a domain-friendly type
// while keeping the implementation details private.
type mongodb[T, C any] Model[
	T,
	C,
	any,
	*options.FindOneOptions,
	*options.FindOptions,
	*options.UpdateOptions,
	mongo.Pipeline,
]

// New creates a new MongoDB model bound to a specific collection.
//
// A single collection instance is reused, which is cheaper than
// resolving the collection on every operation.
func New[T, C any](db *mongo.Database, name string) mongodb[T, C] {
	collection := db.Collection(name)
	return &mongoModel[T, C]{
		Name:       name,
		collection: collection,
	}
}

// FindOne retrieves a single document that matches the given filter.
//
// Decode is used directly into T, avoiding intermediate allocations
// and keeping the operation efficient.
func (m *mongoModel[T, C]) FindOne(
	ctx context.Context,
	filter any,
	options ...*options.FindOneOptions,
) (T, error) {
	var result T
	if err := m.collection.FindOne(ctx, filter, options...).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

// FindMany retrieves all documents that match the given filter.
func (m *mongoModel[T, C]) FindMany(
	ctx context.Context,
	filter any,
	options ...*options.FindOptions,
) ([]T, error) {
	cursor, err := m.collection.Find(ctx, filter, options...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	results := make([]T, 0)

	for cursor.Next(ctx) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}


// Create inserts a new document into the collection.
func (m *mongoModel[T, C]) Create(ctx context.Context, v T) error {
	_, err := m.collection.InsertOne(ctx, v)
	return err
}

// UpdateOne updates a single document that matches the given filter.
func (m *mongoModel[T, C]) UpdateOne(
	ctx context.Context,
	filter any,
	update any,
	options ...*options.UpdateOptions,
) error {
	_, err := m.collection.UpdateOne(ctx, filter, update, options...)
	return err
}

// UpdateMany updates all documents that match the given filter.
func (m *mongoModel[T, C]) UpdateMany(
	ctx context.Context,
	filter any,
	update any,
	options ...*options.UpdateOptions,
) error {
	_, err := m.collection.UpdateMany(ctx, filter, update, options...)
	return err
}

// DeleteOne removes a single document that matches the given filter.
func (m *mongoModel[T, C]) DeleteOne(ctx context.Context, filter any) error {
	_, err := m.collection.DeleteOne(ctx, filter)
	return err
}

// DeleteMany removes all documents that match the given filter.
func (m *mongoModel[T, C]) DeleteMany(ctx context.Context, filter any) error {
	_, err := m.collection.DeleteMany(ctx, filter)
	return err
}

// Aggregate executes an aggregation pipeline and decodes the results into C.
func (m *mongoModel[T, C]) Aggregate(
	ctx context.Context,
	pipeline mongo.Pipeline,
) ([]C, error) {
	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute aggregation: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]C, 0)

	for cursor.Next(ctx) {
		var item C
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode aggregation result: %w", err)
		}
		results = append(results, item)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return results, nil
}


// Model defines a generic interface for database operations.
//
// Generics provide compile-time safety and remove the need for
// interface{} casting, which improves readability and performance.
type Model[T, C, D, FO, FMO, UO, P any] interface {

	// FindOne finds a single document that matches the filter.
	FindOne(ctx context.Context, filter D, options ...FO) (T, error)

	// FindMany finds all documents that match the filter.
	FindMany(ctx context.Context, filter D, options ...FMO) ([]T, error)

	// Create inserts a new document.
	Create(ctx context.Context, data T) error

	// UpdateOne updates a single document that matches the filter.
	UpdateOne(ctx context.Context, filter D, data D, options ...UO) error

	// UpdateMany updates multiple documents that match the filter.
	UpdateMany(ctx context.Context, filter D, data D, options ...UO) error

	// DeleteOne deletes a single document that matches the filter.
	DeleteOne(ctx context.Context, filter D) error

	// DeleteMany deletes all documents that match the filter.
	DeleteMany(ctx context.Context, filter D) error

	// Aggregate executes an aggregation pipeline and returns custom results.
	Aggregate(ctx context.Context, pipeline P) ([]C, error)
}
