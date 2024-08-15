package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/betterstack-community/go-blog/models"
)

type PostRepository struct {
	dbpool *pgxpool.Pool
}

func NewPostRepository(dbpool *pgxpool.Pool) *PostRepository {
	return &PostRepository{dbpool}
}

func (pr *PostRepository) CreatePost(post *models.Post) error {
	query := "INSERT INTO posts (title, content) VALUES ($1, $2) RETURNING id"
	return pr.dbpool.QueryRow(context.Background(), query, post.Title, post.Content).
		Scan(&post.ID)
}

func (pr *PostRepository) GetPosts(ctx context.Context) ([]models.Post, error) {
	rows, err := pr.dbpool.Query(
		ctx,
		"SELECT id, title, content FROM posts",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (pr *PostRepository) GetPost(id int) (*models.Post, error) {
	var post models.Post
	query := "SELECT id, title, content FROM posts WHERE id = $1"
	err := pr.dbpool.QueryRow(context.Background(), query, id).
		Scan(&post.ID, &post.Title, &post.Content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Or a custom "not found" error
		}
		return nil, err
	}

	return &post, nil
}

func (pr *PostRepository) UpdatePost(post *models.Post) error {
	query := "UPDATE posts SET title = $1, content = $2 WHERE id = $3"
	_, err := pr.dbpool.Exec(
		context.Background(),
		query,
		post.Title,
		post.Content,
		post.ID,
	)
	return err
}

func (pr *PostRepository) DeletePost(id int) error {
	query := "DELETE FROM posts WHERE id = $1"
	_, err := pr.dbpool.Exec(context.Background(), query, id)
	return err
}
