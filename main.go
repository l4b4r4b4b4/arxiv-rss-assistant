package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/l4b4r4b4b4/arxiv-rss-assistant/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT environment variable was not set")
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable was not set")
	}
	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Can't connect to database", err)
	}
	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}

	go startScraping(db, 10, time.Minute)

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", readinessHandler)
	v1Router.Get("/err", errorHandler)
	v1Router.Post("/users", apiCfg.createUserHandler)
	v1Router.Get("/users", apiCfg.authMiddleware(apiCfg.getUserHandler))
	v1Router.Post("/feeds", apiCfg.authMiddleware(apiCfg.createFeedHandler))
	v1Router.Get("/feeds", apiCfg.getFeedsHandler)
	v1Router.Post("/feed_follows", apiCfg.authMiddleware(apiCfg.createFeedFollowHandler))
	v1Router.Get("/feed_follows", apiCfg.authMiddleware(apiCfg.getFeedFollowsHandler))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.authMiddleware(apiCfg.deleteFeedFollowsHandler))
	v1Router.Get("/posts", apiCfg.authMiddleware(apiCfg.getPostsForUserHandler))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}
	log.Printf("Server started on port %s", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
