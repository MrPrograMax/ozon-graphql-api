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

func TestMemoryCreatePost_Success(t *testing.T) {
	storage := memory.NewStorage()

	storage.Users["1"] = &model.User{
		ID:       "1",
		Username: "Maxim",
	}

	repo := &repository.MemoryPostRepository{Storage: storage}

	input := model.NewPost{
		Title:                 "New Post",
		Text:                  "This is a new post.",
		UserID:                "1",
		IsCommentingAvailable: boolPtr(true),
	}

	post, err := repo.CreatePost(context.Background(), input)
	require.NoError(t, err)

	expectedPost := &model.Post{
		ID:                    "1",
		Title:                 "New Post",
		Text:                  "This is a new post.",
		CreatedBy:             &model.User{ID: "1"},
		CreatedAt:             post.CreatedAt,
		IsCommentingAvailable: true,
		Comments:              []*model.Comment{},
	}

	assert.Equal(t, expectedPost.Title, post.Title)
	assert.Equal(t, expectedPost.Text, post.Text)
	assert.Equal(t, expectedPost.CreatedBy.ID, post.CreatedBy.ID)
	assert.True(t, len(post.CreatedAt) > 0)

	storedPost, exists := storage.Posts[post.ID]
	require.True(t, exists)
	assert.Equal(t, post, storedPost)
}

func TestCreatePost_UserNotFound(t *testing.T) {
	storage := memory.NewStorage()

	repo := &repository.MemoryPostRepository{Storage: storage}

	input := model.NewPost{
		Title:                 "New Post",
		Text:                  "This is a new post.",
		UserID:                "nonexistentUser",
		IsCommentingAvailable: boolPtr(true),
	}

	post, err := repo.CreatePost(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "user doesn't exists", err.Error())
}

func boolPtr(b bool) *bool {
	return &b
}
