// Package mongodb provides a generic MongoDB model implementation
package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	*options.UpdateOneOptions,
	*options.UpdateManyOptions,
	mongo.Pipeline,
]

// DefaultModel is the default MongoDB model type alias.
type DefaultModel[T, C any] = mongodb[T, C]

// New creates a new MongoDB model bound to a specific collection.
//
// A single collection instance is reused, which is cheaper than
// resolving the collection on every operation.
func New[T, C any](db *mongo.Database, name string) DefaultModel[T, C] {
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
	opts ...*options.FindOneOptions,
) (T, error) {
	var result T
	findOneOpts := options.FindOne()
	findOneOpts.Opts = []func(*options.FindOneOptions) error{
		setOptions(opts...),
	}
	if err := m.collection.FindOne(ctx, filter, findOneOpts).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

// FindMany retrieves all documents that match the given filter.
func (m *mongoModel[T, C]) FindMany(
	ctx context.Context,
	filter any,
	opts ...*options.FindOptions,
) ([]T, error) {
	findOpts := options.Find()
	findOpts.Opts = []func(*options.FindOptions) error{
		setOptions(opts...),
	}
	cursor, err := m.collection.Find(ctx, filter, findOpts)
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
	opts ...*options.UpdateOneOptions,
) error {
	updateOneOpts := options.UpdateOne()
	updateOneOpts.Opts = []func(*options.UpdateOneOptions) error{
		setOptions(opts...),
	}
	_, err := m.collection.UpdateOne(ctx, filter, update, updateOneOpts)
	return err
}

// UpdateMany updates all documents that match the given filter.
func (m *mongoModel[T, C]) UpdateMany(
	ctx context.Context,
	filter any,
	update any,
	opts ...*options.UpdateManyOptions,
) error {
	updateManyOpts := options.UpdateMany()
	updateManyOpts.Opts = []func(*options.UpdateManyOptions) error{
		setOptions(opts...),
	}
	_, err := m.collection.UpdateMany(ctx, filter, update, updateManyOpts)
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

func setOptions[T any](opts ...*T) func(opts *T) error {
	fn := func(o *T) error {
		if len(opts) > 0 {
			o = opts[0]
		}
		return nil
	}
	return fn
}

// Model defines a generic interface for database operations.
//
// Generics provide compile-time safety and remove the need for
// interface{} casting, which improves readability and performance.
type Model[T, C, D, FO, FMO, UO, UM, P any] interface {

	// FindOne finds a single document that matches the filter.
	FindOne(ctx context.Context, filter D, options ...FO) (T, error)

	// FindMany finds all documents that match the filter.
	FindMany(ctx context.Context, filter D, options ...FMO) ([]T, error)

	// Create inserts a new document.
	Create(ctx context.Context, data T) error

	// UpdateOne updates a single document that matches the filter.
	UpdateOne(ctx context.Context, filter D, data D, options ...UO) error

	// UpdateMany updates multiple documents that match the filter.
	UpdateMany(ctx context.Context, filter D, data D, options ...UM) error

	// DeleteOne deletes a single document that matches the filter.
	DeleteOne(ctx context.Context, filter D) error

	// DeleteMany deletes all documents that match the filter.
	DeleteMany(ctx context.Context, filter D) error

	// Aggregate executes an aggregation pipeline and returns custom results.
	Aggregate(ctx context.Context, pipeline P) ([]C, error)
}
