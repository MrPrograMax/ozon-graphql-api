package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ozon-graphql-api/graph"
	"ozon-graphql-api/internal/repository"
	"ozon-graphql-api/pkg/database"
	"ozon-graphql-api/pkg/memory"
	"sync"
	"time"
)

const defaultPort = "8080"

const (
	CONFIG_DIR  = "configs"
	CONFIG_FILE = "config"
)

func initConfig() error {
	viper.AddConfigPath(CONFIG_DIR)
	viper.SetConfigName(CONFIG_FILE)
	return viper.ReadInConfig()
}

func LoadFromFile(filename string) (*memory.Storage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var s memory.Storage
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func main() {
	var useMemoryStorage bool
	flag.BoolVar(&useMemoryStorage, "m", false, "Use in-memory storage")
	flag.Parse()

	if err := initConfig(); err != nil {
		log.Println(err)
		return
	}

	if err := godotenv.Load("/app/.env"); err != nil {
		log.Println(err)
		return
	}

	var repos *repository.Repository
	var server *http.Server
	var storage *memory.Storage

	if useMemoryStorage {
		log.Println("Service started with using storage")
		if s, err := LoadFromFile("storage.json"); err == nil {
			storage = s
		} else {
			storage = memory.NewStorage()
		}

		repos = repository.NewMemoryRepository(storage)
	} else {
		log.Println("Service started with using db ")
		db, err := database.NewPostgresDB(database.Config{
			Host:     viper.GetString("db.host"),
			Port:     viper.GetString("db.port"),
			Username: viper.GetString("db.username"),
			DBName:   viper.GetString("db.dbname"),
			SSLMode:  viper.GetString("db.sslmode"),
			Password: os.Getenv("DB_PASSWORD"),
		})
		if err != nil {
			log.Println(err)
			return
		}
		repos = repository.NewPostgresRepository(db)
	}

	port := viper.GetString("http.port")
	if port == "" {
		port = defaultPort
	}

	resolver := graph.NewResolver(repos)
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	wg := &sync.WaitGroup{}

	server = &http.Server{Addr: ":" + port}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe() error: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	log.Println("Shutting down...")

	if useMemoryStorage && storage != nil {
		if err := storage.SaveToFile("storage.json"); err != nil {
			log.Printf("Error saving storage to file: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server stopped")

	wg.Wait()
}
