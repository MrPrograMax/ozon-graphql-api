package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/internal/repository"
	"ozon-graphql-api/pkg/memory"
	"testing"
)

func TestComments(t *testing.T) {
	storage := memory.NewStorage()

	storage.Comments["1"] = &model.Comment{
		ID:        "1",
		PostID:    "post1",
		Sender:    &model.User{ID: "1"},
		ReplyTo:   nil,
		Text:      "Comment 1",
		CreatedAt: "2024-09-10T10:00:00Z",
	}
	storage.Comments["2"] = &model.Comment{
		ID:        "2",
		PostID:    "post1",
		Sender:    &model.User{ID: "2"},
		ReplyTo:   nil,
		Text:      "Comment 2",
		CreatedAt: "2024-09-11T10:00:00Z",
	}

	repo := &repository.MemoryCommentRepository{Storage: storage}

	limit := 2
	offset := 0

	comments, err := repo.Comments(context.Background(), &limit, &offset)
	require.NoError(t, err)

	expectedComments := []*model.Comment{
		{
			ID:        "2",
			PostID:    "post1",
			Sender:    &model.User{ID: "2"},
			ReplyTo:   nil,
			Text:      "Comment 2",
			CreatedAt: "2024-09-11T10:00:00Z",
		},
		{
			ID:        "1",
			PostID:    "post1",
			Sender:    &model.User{ID: "1"},
			ReplyTo:   nil,
			Text:      "Comment 1",
			CreatedAt: "2024-09-10T10:00:00Z",
		},
	}

	assert.ElementsMatch(t, expectedComments, comments)
}

func TestMemoryCreateComment_Success(t *testing.T) {
	storage := memory.NewStorage()

	storage.Posts["post1"] = &model.Post{
		ID:                    "post1",
		IsCommentingAvailable: true,
	}
	storage.Users["4"] = &model.User{ID: "4"}

	repo := &repository.MemoryCommentRepository{Storage: storage}

	input := model.NewComment{
		PostID:   "post1",
		SenderID: "4",
		Text:     "This is a comment",
		ReplyTo:  nil,
	}

	comment, err := repo.CreateComment(context.Background(), input)
	require.NoError(t, err)

	expectedComment := &model.Comment{
		ID:        "1",
		PostID:    "post1",
		Sender:    &model.User{ID: "4"},
		ReplyTo:   nil,
		Text:      "This is a comment",
		CreatedAt: comment.CreatedAt,
	}

	assert.Equal(t, expectedComment.PostID, comment.PostID)
	assert.Equal(t, expectedComment.Sender.ID, comment.Sender.ID)
	assert.Equal(t, expectedComment.Text, comment.Text)
	assert.True(t, len(comment.CreatedAt) > 0)
}

func TestCreateComment_PostNotFound(t *testing.T) {
	storage := memory.NewStorage()

	repo := &repository.MemoryCommentRepository{Storage: storage}

	input := model.NewComment{
		PostID:   "nonexistentPost",
		SenderID: "1",
		Text:     "This comment should fail",
	}

	comment, err := repo.CreateComment(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "post not found", err.Error())
}

func TestCreateComment_CommentingNotAllowed(t *testing.T) {
	storage := memory.NewStorage()
	storage.Posts["post1"] = &model.Post{
		ID:                    "post1",
		IsCommentingAvailable: false,
	}

	repo := &repository.MemoryCommentRepository{Storage: storage}

	input := model.NewComment{
		PostID:   "post1",
		SenderID: "1",
		Text:     "This comment should fail",
	}

	comment, err := repo.CreateComment(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "commenting is not allowed on this post", err.Error())
}

func TestCreateComment_SenderNotFound(t *testing.T) {
	storage := memory.NewStorage()
	storage.Posts["post1"] = &model.Post{
		ID:                    "post1",
		IsCommentingAvailable: true,
	}

	repo := &repository.MemoryCommentRepository{Storage: storage}

	input := model.NewComment{
		PostID:   "post1",
		SenderID: "nonexistentUser",
		Text:     "This comment should fail",
	}

	comment, err := repo.CreateComment(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "sender not found", err.Error())
}
