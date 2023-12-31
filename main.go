package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/2twin/blog/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	// feed, err := urlToFeed("https://wagslane.dev/index.xml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(feed)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't find .env file")
	}

	const filepathRoot = "."
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	apiCfg := apiConfig{
		DB: dbQueries,
	}

	go startScraping(dbQueries, 10, time.Minute)

	router := chi.NewRouter()
	router.Use(cors.Handler(
		cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		},
	))

	apiRouter := chi.NewRouter()
	apiRouter.Get("/readiness", handlerReadiness)
	apiRouter.Get("/err", handlerError)

	apiRouter.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	apiRouter.Post("/users", apiCfg.handlerCreateUser)

	apiRouter.Get("/feeds", apiCfg.handlerGetFeeds)
	apiRouter.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))

	apiRouter.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	apiRouter.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	apiRouter.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	apiRouter.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsByUser))

	router.Mount("/v1", apiRouter)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
