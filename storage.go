package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ctxTimeout = 5 * time.Second
)

type storage interface {
	initialize() error
	shutdown() error

	get(id int) (*todo, error)
	getAll() ([]*todo, error)
	contains(id int) (bool, error)

	add(*todo) error
	delete(id int) error
}

type inMemoryStore struct {
	store map[int]*todo
}

func (s *inMemoryStore) initialize() error {
	s.store = make(map[int]*todo, 0)
	return nil
}

func (s *inMemoryStore) shutdown() error {
	s.store = nil
	return nil
}

func (s *inMemoryStore) get(id int) (*todo, error) {
	t, ok := s.store[id]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	return t, nil
}

func (s *inMemoryStore) getAll() ([]*todo, error) {
	t := []*todo{}
	for _, v := range s.store {
		t = append(t, v)
	}
	return t, nil
}

func (s *inMemoryStore) contains(id int) (bool, error) {
	_, ok := s.store[id]
	return ok, nil
}

func (s *inMemoryStore) add(t *todo) error {
	s.store[t.Id] = t
	return nil
}

func (s *inMemoryStore) delete(id int) error {
	delete(s.store, id)
	return nil
}

type mongoStore struct {
	ctx        context.Context
	uri        string
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func (s *mongoStore) initialize() error {
	s.ctx = context.Background()
	s.uri = parseMongoURI()

	client, err := mongo.NewClient(options.Client().ApplyURI(s.uri))
	if err != nil {
		return err
	}
	s.client = client

	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	if err = s.client.Connect(ctx); err != nil {
		return err
	}

	if err = s.client.Ping(ctx, nil); err != nil {
		return err
	}

	db := os.Getenv("MONGO_DATABASE")
	if db == "" {
		db = "todo_db"
	}
	coll := os.Getenv("MONGO_COLLECTION")
	if coll == "" {
		coll = "todos"
	}

	s.database = s.client.Database(db)
	s.collection = s.database.Collection(coll)

	return nil
}

func (s *mongoStore) shutdown() error {
	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	err := s.client.Disconnect(ctx)
	return err
}

func (s *mongoStore) get(id int) (*todo, error) {
	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	result := s.collection.FindOne(ctx, bson.M{"id": id})
	t := &todo{}
	err := result.Decode(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *mongoStore) getAll() ([]*todo, error) {
	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	t := &[]*todo{}
	err = cursor.All(ctx, t)

	return *t, err
}

func (s *mongoStore) contains(id int) (bool, error) {
	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	count, err := s.collection.CountDocuments(ctx, bson.M{"id": id})
	if err != nil {
		return false, err
	}
	found := count > 0
	return found, nil

}

func (s *mongoStore) add(t *todo) error {
	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	_, err := s.collection.InsertOne(ctx, t)
	return err
}

func (s *mongoStore) delete(id int) error {
	ctx, cancel := context.WithTimeout(s.ctx, ctxTimeout)
	defer cancel()

	_, err := s.collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func parseMongoURI() string {
	defaultURI := "mongodb://root:password@localhost:27017"

	user := os.Getenv("MONGO_USER")
	password := os.Getenv("MONGO_PASSWORD")
	host := os.Getenv("MONGO_HOST")
	port := os.Getenv("MONGO_PORT")

	if user == "" || password == "" || host == "" || port == "" {
		return defaultURI
	}
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port)
}
