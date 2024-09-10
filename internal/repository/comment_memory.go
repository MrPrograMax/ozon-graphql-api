package repository

import (
	"context"
	"errors"
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/pkg/memory"
	"sort"
	"strconv"
	"time"
)

type MemoryCommentRepository struct {
	Storage *memory.Storage
}

func NewMemoryCommentRepo(storage *memory.Storage) *MemoryCommentRepository {
	return &MemoryCommentRepository{
		Storage: storage,
	}
}

func (r *MemoryCommentRepository) Comments(ctx context.Context, limit, offset *int) ([]*model.Comment, error) {
	r.Storage.Mu.RLock()
	defer r.Storage.Mu.RUnlock()

	var comments []*model.Comment
	for _, comment := range r.Storage.Comments {
		comments = append(comments, comment)
	}

	//Т.к. обход мап в гошке случайный, отсортируем нашу мапу по дате добавления для красивого вывода
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt > comments[j].CreatedAt
	})

	start := *offset
	end := start + *limit

	if end > len(comments) {
		end = len(comments)
	}

	if start > len(comments) {
		return []*model.Comment{}, nil
	}

	return comments[start:end], nil
}

func (r *MemoryCommentRepository) CreateComment(ctx context.Context, input model.NewComment) (*model.Comment, error) {
	r.Storage.Mu.RLock()
	post, ok := r.Storage.Posts[input.PostID]
	r.Storage.Mu.RUnlock()
	if !ok {
		return nil, errors.New("post not found")
	}

	if !post.IsCommentingAvailable {
		return nil, errors.New("commenting is not allowed on this post")
	}

	r.Storage.Mu.RLock()
	sender, ok := r.Storage.Users[input.SenderID]
	r.Storage.Mu.RUnlock()
	if !ok {
		return nil, errors.New("sender not found")
	}

	r.Storage.Mu.Lock()
	r.Storage.CommentIdCounter++
	commentId := strconv.Itoa(r.Storage.CommentIdCounter)
	r.Storage.Mu.Unlock()

	newComment := &model.Comment{
		ID:        commentId,
		PostID:    input.PostID,
		Sender:    sender,
		ReplyTo:   nil,
		Text:      input.Text,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if input.ReplyTo != nil {
		comment, ok := r.Storage.Comments[*input.ReplyTo]
		if !ok {
			return nil, errors.New("comment you want to reply to doesn't exist")
		}

		newComment.ReplyTo = &model.Comment{
			ID: *input.ReplyTo,
		}

		comment.Replies = append(comment.Replies, newComment)
	}

	r.Storage.Mu.Lock()
	r.Storage.Comments[commentId] = newComment
	r.Storage.Mu.Unlock()

	return newComment, nil
}
