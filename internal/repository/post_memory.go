package repository

import (
	"context"
	"errors"
	"fmt"
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/pkg/memory"
	"sort"
	"strconv"
	"time"
)

type MemoryPostRepository struct {
	Storage *memory.Storage
}

func NewMemoryPostRepo(storage *memory.Storage) *MemoryPostRepository {
	return &MemoryPostRepository{
		Storage: storage,
	}
}

func (r *MemoryPostRepository) Posts(ctx context.Context, limit, offset *int) ([]*model.Post, error) {
	r.Storage.Mu.RLock()
	defer r.Storage.Mu.RUnlock()

	var posts []*model.Post
	for _, post := range r.Storage.Posts {
		posts = append(posts, post)
	}

	//Т.к. обход мап в гошке случайный, отсортируем нашу мапу по дате добавления для красивого вывода
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt > posts[j].CreatedAt
	})

	start := *offset
	end := start + *limit

	if end > len(posts) {
		end = len(posts)
	}

	if start > len(posts) {
		return []*model.Post{}, nil
	}

	return posts[start:end], nil
}

func (r *MemoryPostRepository) PostByID(ctx context.Context, id int) (*model.Post, error) {
	r.Storage.Mu.RLock()
	defer r.Storage.Mu.RUnlock()

	post, exists := r.Storage.Posts[strconv.Itoa(id)]
	if !exists {
		return nil, fmt.Errorf("post with id %d not found", id)
	}

	var rootComments []*model.Comment
	for _, comment := range r.Storage.Comments {
		if comment.PostID == strconv.Itoa(id) && comment.ReplyTo == nil {
			rootComments = append(rootComments, comment)
		}
	}

	resultPost := &model.Post{
		ID:                    post.ID,
		Title:                 post.Title,
		Text:                  post.Text,
		CreatedAt:             post.CreatedAt,
		IsCommentingAvailable: post.IsCommentingAvailable,
		CreatedBy: &model.User{
			ID:       post.CreatedBy.ID,
			Username: post.CreatedBy.Username,
		},
		Comments: rootComments,
	}

	return resultPost, nil
}

func (r *MemoryPostRepository) CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error) {
	//Проверка существования пользователя
	r.Storage.Mu.RLock()
	user, ok := r.Storage.Users[input.UserID]
	r.Storage.Mu.RUnlock()
	if !ok {
		return nil, errors.New("user doesn't exists")
	}

	r.Storage.Mu.Lock()
	postId := r.Storage.PostIdCounter + 1
	r.Storage.PostIdCounter = postId
	r.Storage.Mu.Unlock()

	newPost := &model.Post{
		ID:                    strconv.Itoa(postId),
		Title:                 input.Title,
		Text:                  input.Text,
		CreatedBy:             user,
		CreatedAt:             time.Now().Format(time.RFC3339),
		IsCommentingAvailable: *input.IsCommentingAvailable,
		Comments:              []*model.Comment{},
	}

	r.Storage.Mu.Lock()
	r.Storage.Posts[strconv.Itoa(postId)] = newPost
	r.Storage.Mu.Unlock()

	return newPost, nil
}
