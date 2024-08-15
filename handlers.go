package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/betterstack-community/go-blog/models"
)

func createPost(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newPost := models.Post{
		Title:   strings.TrimSpace(r.FormValue("title")),
		Content: strings.TrimSpace(r.FormValue("content")),
	}

	if newPost.Title == "" || newPost.Content == "" {
		http.Error(w, "Title or content cannot be empty", http.StatusBadRequest)
		return
	}

	if err := postRepo.CreatePost(&newPost); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getPosts(
	w http.ResponseWriter,
	r *http.Request,
) {
	posts, err := postRepo.GetPosts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFS(
		templates,
		filepath.Join(
			"templates",
			"default.html",
		),
		filepath.Join("templates", "index.html"),
	))

	tmpl.ExecuteTemplate(w, "default", struct{ Posts []models.Post }{posts})
}

func getPost(
	r *http.Request,
) (*models.Post, error) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("Invalid post ID: %d", id)
	}

	post, err := postRepo.GetPost(id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, fmt.Errorf("Post not found")
	}

	return post, nil
}

func updatePost(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedPost := models.Post{
		ID:      id,
		Title:   strings.TrimSpace(r.FormValue("title")),
		Content: strings.TrimSpace(r.FormValue("content")),
	}

	if updatedPost.Title == "" || updatedPost.Content == "" {
		http.Error(w, "Title or content cannot be empty", http.StatusBadRequest)
		return
	}

	if err := postRepo.UpdatePost(&updatedPost); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deletePost(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if err := postRepo.DeletePost(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
