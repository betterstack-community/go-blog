package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/betterstack-community/go-blog/models"
	"github.com/betterstack-community/go-blog/repository"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

//go:embed templates/*
var templates embed.FS

var postRepo *repository.PostRepository

type Config struct {
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresURL      string
}

var conf Config

var postFormTmpl = template.Must(template.ParseFS(
	templates,
	filepath.Join(
		"templates",
		"default.html",
	),
	filepath.Join("templates", "create-post.html"),
))

func init() {
	godotenv.Load()

	conf.PostgresDB = os.Getenv("POSTGRES_DB")
	conf.PostgresUser = os.Getenv("POSTGRES_USER")
	conf.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	conf.PostgresHost = os.Getenv("POSTGRES_HOST")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		conf.PostgresUser,
		conf.PostgresPassword,
		conf.PostgresHost,
		conf.PostgresDB,
	)

	conf.PostgresURL = connStr
}

// Database connection using pgxpool.
func main() {
	ctx := context.Background()

	dbpool, err := pgxpool.New(ctx, conf.PostgresURL)
	if err != nil {
		log.Fatal("Unable to create database connection pool:", err)
	}

	defer dbpool.Close()

	err = dbpool.Ping(ctx)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	postRepo = repository.NewPostRepository(dbpool)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getPosts(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts/new", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			createPost(w, r)
		case "GET":
			postFormTmpl.ExecuteTemplate(
				w,
				"default",
				struct{ Post models.Post }{},
			)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc(
		"/post/{id}/edit",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				post, err := getPost(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				postFormTmpl.ExecuteTemplate(
					w,
					"default",
					struct{ Post *models.Post }{post},
				)
			case "POST":
				updatePost(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		},
	)

	http.HandleFunc(
		"/post/{id}/delete",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				deletePost(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		},
	)

	http.HandleFunc("/post/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			post, err := getPost(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tmpl := template.Must(template.ParseFS(
				templates,
				filepath.Join(
					"templates",
					"default.html",
				),
				filepath.Join("templates", "post.html"),
			))

			tmpl.ExecuteTemplate(
				w,
				"default",
				struct{ Post *models.Post }{post},
			)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
