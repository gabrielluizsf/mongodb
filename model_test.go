package mongodb

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type testUser struct {
	ID       string `bson:"_id,omitempty"`
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Age      int    `bson:"age"`
	Position string `bson:"position"`
}

type testEmployee struct {
	ID        string `bson:"_id,omitempty"`
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Position  string `bson:"position"`
}

func TestMongoModel(t *testing.T) {
	ctx := context.Background()
	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("DATABASE_NAME")

	if uri == "" || dbName == "" {
		t.Skip("env not set")
	}
	db, err := NewConnector(dbName, uri).Connect()
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	_ = db.Collection("users").Drop(ctx)

	model := New[testUser, testEmployee](db, "users")

	t.Run("Create", func(t *testing.T) {
		err := model.Create(ctx, testUser{
			ID:       "1",
			Name:     "Alice",
			Email:    "alice@test.com",
			Age:      30,
			Position: "Dev",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = model.Create(ctx, testUser{
			ID:       "2",
			Name:     "Bob",
			Email:    "bob@test.com",
			Age:      35,
			Position: "QA",
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FindMany", func(t *testing.T) {
		users, err := model.FindMany(ctx, map[string]any{})
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 {
			t.Fatalf("expected 2, got %d", len(users))
		}
	})

	t.Run("FindOne", func(t *testing.T) {
		user, err := model.FindOne(ctx, map[string]any{"email": "alice@test.com"})
		if err != nil {
			t.Fatal(err)
		}
		if user.Name != "Alice" {
			t.Fatalf("unexpected user %+v", user)
		}
	})

	t.Run("UpdateOne", func(t *testing.T) {
		err := model.UpdateOne(
			ctx,
			map[string]any{"email": "alice@test.com"},
			map[string]any{"$set": map[string]any{"age": 31}},
		)
		if err != nil {
			t.Fatal(err)
		}

		user, _ := model.FindOne(ctx, map[string]any{"email": "alice@test.com"})
		if user.Age != 31 {
			t.Fatalf("expected 31, got %d", user.Age)
		}
	})

	t.Run("UpdateMany", func(t *testing.T) {
		err := model.UpdateMany(
			ctx,
			map[string]any{"position": "QA"},
			map[string]any{"$set": map[string]any{"position": "Tester"}},
		)
		if err != nil {
			t.Fatal(err)
		}

		user, _ := model.FindOne(ctx, map[string]any{"email": "bob@test.com"})
		if user.Position != "Tester" {
			t.Fatalf("unexpected position %s", user.Position)
		}
	})

	t.Run("Aggregate", func(t *testing.T) {
		results, err := model.Aggregate(ctx, mongo.Pipeline{
			{{Key: "$project", Value: map[string]any{
				"first_name": "$name",
				"position":   1,
			}}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) == 0 {
			t.Fatal("empty aggregation")
		}
	})

	t.Run("DeleteOne", func(t *testing.T) {
		err := model.DeleteOne(ctx, map[string]any{"email": "alice@test.com"})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DeleteMany", func(t *testing.T) {
		err := model.DeleteMany(ctx, map[string]any{})
		if err != nil {
			t.Fatal(err)
		}

		users, _ := model.FindMany(ctx, map[string]any{})
		if len(users) != 0 {
			t.Fatalf("expected empty collection")
		}
	})
}
