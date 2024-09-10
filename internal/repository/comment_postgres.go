package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"ozon-graphql-api/graph/model"
	"strconv"
)

type PostgresCommentRepository struct {
	Db *sqlx.DB
}

func NewPostgresCommentRepo(db *sqlx.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{
		Db: db,
	}
}

func (r *PostgresCommentRepository) Comments(ctx context.Context, limit, offset *int) ([]*model.Comment, error) {
	queryLimit := *limit
	queryOffset := *offset

	commentFields := `c.id, c.postid, c.sender, c.replyto, c.text, c.createdat, u.id, u.username`

	query := fmt.Sprintf(`SELECT %s FROM %s c JOIN %s u on c.sender = u.id 
                              ORDER BY createdat DESC LIMIT $1 OFFSET $2`,
		commentFields, commentsTable, usersTable)

	// Промежуточная структура для маппинга
	var dbComments []dbCommentStruct
	err := r.Db.SelectContext(ctx, &dbComments, query, queryLimit, queryOffset)
	if err != nil {
		return nil, err
	}

	var comments []*model.Comment
	for _, c := range dbComments {
		comment := &model.Comment{
			ID:     strconv.Itoa(c.ID),
			PostID: strconv.Itoa(c.PostID),
			Sender: &model.User{
				ID:       strconv.Itoa(c.SenderID),
				Username: *c.Username,
			},
			ReplyTo:   nil,
			Text:      c.Text,
			CreatedAt: c.CreatedAt,
		}

		if c.ReplyTo != nil {
			comment.ReplyTo = &model.Comment{
				ID: strconv.Itoa(*c.ReplyTo),
			}
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *PostgresCommentRepository) CreateComment(ctx context.Context, input model.NewComment) (*model.Comment, error) {
	//Ограничение на размерность текста сообщения
	if len([]rune(input.Text)) > 2000 {
		return nil, errors.New("text should not exceed 2000 characters")
	}

	var isCommentingAvailable bool
	postQuery := fmt.Sprintf(`SELECT isCommentingAvailable FROM %s WHERE id = $1`, postsTable)
	err := r.Db.QueryRowContext(ctx, postQuery, input.PostID).Scan(&isCommentingAvailable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	if !isCommentingAvailable {
		return nil, errors.New("commenting is not allowed on this post")
	}

	var query string
	var commentId int
	var createdAt string

	fieldsWithNull := `postid, sender, text`
	fields := `postid, sender, replyto, text`

	if input.ReplyTo == nil {
		query = fmt.Sprintf(`INSERT INTO %s (%s) VALUES ($1, $2, $3) RETURNING id, createdAt`, commentsTable, fieldsWithNull)
		err := r.Db.QueryRowContext(ctx, query, input.PostID, input.SenderID, input.Text).Scan(&commentId, &createdAt)
		if err != nil {
			return nil, err
		}

		comment := &model.Comment{
			ID:     strconv.Itoa(commentId),
			PostID: input.PostID,
			Sender: &model.User{
				ID: input.SenderID,
			},
			ReplyTo:   nil,
			Text:      input.Text,
			CreatedAt: createdAt,
		}

		return comment, nil
	}

	query = fmt.Sprintf(`INSERT INTO %s (%s) VALUES ($1, $2, $3, $4) RETURNING id, createdAt`, commentsTable, fields)
	err = r.Db.QueryRowContext(ctx, query, input.PostID, input.SenderID, input.ReplyTo, input.Text).Scan(&commentId, &createdAt)
	if err != nil {
		return nil, err
	}

	comment := &model.Comment{
		ID:     strconv.Itoa(commentId),
		PostID: input.PostID,
		Sender: &model.User{
			ID: input.SenderID,
		},
		ReplyTo: &model.Comment{
			ID: *input.ReplyTo,
		},
		Text:      input.Text,
		CreatedAt: createdAt,
	}

	return comment, nil
}
