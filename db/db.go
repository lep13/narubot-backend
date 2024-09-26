package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"encoding/json"
	"narubot-backend/models" // Import the models package to use Config

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClientInterface defines the interface for MongoDB client methods used in our code.
type MongoClientInterface interface {
	Ping(ctx context.Context, rp *readpref.ReadPref) error
	Database(name string, opts ...*options.DatabaseOptions) *mongo.Database
}

// MongoClientWrapper wraps the actual MongoDB client to conform to our interface.
type MongoClientWrapper struct {
	Client *mongo.Client
}

func (m *MongoClientWrapper) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return m.Client.Ping(ctx, rp)
}

func (m *MongoClientWrapper) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	return m.Client.Database(name, opts...)
}

// MongoClient holds the actual MongoDB client or a mock for testing.
var MongoClient MongoClientInterface

// MongoConnectFuncType defines the function type for connecting to MongoDB.
type MongoConnectFuncType func(ctx context.Context, uri string) (MongoClientInterface, error)

// DefaultMongoConnectFunc is the default function for connecting to MongoDB.
var DefaultMongoConnectFunc MongoConnectFuncType = func(ctx context.Context, uri string) (MongoClientInterface, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	return &MongoClientWrapper{Client: client}, nil
}

// CollectionGetterFunc is a function type for getting a collection.
type CollectionGetterFunc func() CollectionInterface

// GetCollectionFunc is a package-level variable holding the function to get a collection.
var GetCollectionFunc CollectionGetterFunc = defaultGetCollection

// CollectionInterface defines the methods to be mocked for MongoDB collection.
type CollectionInterface interface {
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

// defaultGetCollection returns the default collection for chatbot data.
func defaultGetCollection() CollectionInterface {
	return MongoClient.Database("narubot").Collection("chatbot")
}

// GetCollection returns a collection from the MongoDB database.
func GetCollection() CollectionInterface {
	return GetCollectionFunc()
}

// InitializeMongoDB initializes the MongoDB client connection using the MongoDB URI from secrets.
func InitializeMongoDB(config *models.Config) error {
	var err error
	MongoClient, err = DefaultMongoConnectFunc(context.Background(), config.MongoURI)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = MongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return err
	}

	log.Println("Connected to MongoDB successfully")
	return nil
}

// SecretsManagerInterface defines the interface for Secrets Manager client methods used in our code.
type SecretsManagerInterface interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretManagerFunc allows for injecting a custom Secrets Manager function for testing.
var SecretManagerFunc = func() (SecretsManagerInterface, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return secretsmanager.NewFromConfig(cfg), nil
}

// FetchMongoURIFromSecrets fetches the MongoDB URI from AWS Secrets Manager.
func FetchMongoURIFromSecrets(secretName string) (string, error) {
	svc, err := SecretManagerFunc()
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}

	var secretData map[string]interface{}
	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	uri, ok := secretData["MONGO_URI"].(string)
	if !ok {
		return "", fmt.Errorf("MONGO_URI not found in secrets")
	}

	return uri, nil
}
