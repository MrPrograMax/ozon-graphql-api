package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/pkg/memory"
)

type dbPostStruct struct {
	ID                    int     `db:"id"`
	Title                 string  `db:"title"`
	Text                  string  `db:"text"`
	CreatedAt             string  `db:"createdat"`
	IsCommentingAvailable bool    `db:"iscommentingavailable"`
	UserID                *int    `db:"userid"`
	Username              *string `db:"username"`
}

type dbCommentStruct struct {
	ID        int     `db:"id"`
	PostID    int     `db:"postid"`
	Text      string  `db:"text"`
	ReplyTo   *int    `db:"replyto"`
	SenderID  int     `db:"sender"`
	CreatedAt string  `db:"createdat"`
	UserID    *int    `db:"userid"`
	Username  *string `db:"username"`
}

type PostRepository interface {
	Posts(ctx context.Context, limit, offset *int) ([]*model.Post, error)
	PostByID(ctx context.Context, id int) (*model.Post, error)
	CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error)
}

type CommentRepository interface {
	Comments(ctx context.Context, limit, offset *int) ([]*model.Comment, error)
	CreateComment(ctx context.Context, input model.NewComment) (*model.Comment, error)
}

type Repository struct {
	PostRepository
	CommentRepository
}

func NewPostgresRepository(db *sqlx.DB) *Repository {
	return &Repository{
		PostRepository:    NewPostgresPostRepo(db),
		CommentRepository: NewPostgresCommentRepo(db),
	}
}

func NewMemoryRepository(storage *memory.Storage) *Repository {
	return &Repository{
		PostRepository:    NewMemoryPostRepo(storage),
		CommentRepository: NewMemoryCommentRepo(storage),
	}
}
