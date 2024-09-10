package test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/internal/repository"
	"testing"
)

func TestCreateComment_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresCommentRepository{Db: sqlxDB}

	input := model.NewComment{
		PostID:   "101",
		SenderID: "1",
		Text:     "This is a comment",
		ReplyTo:  nil,
	}

	mock.ExpectQuery(`^SELECT isCommentingAvailable FROM posts WHERE id = \$1$`).
		WithArgs(input.PostID).
		WillReturnRows(sqlmock.NewRows([]string{"isCommentingAvailable"}).AddRow(true))

	insertQuery := fmt.Sprintf(`^INSERT INTO comments \(postid, sender, text\) VALUES \(\$1, \$2, \$3\) RETURNING id, createdAt$`)
	mock.ExpectQuery(insertQuery).
		WithArgs(input.PostID, input.SenderID, input.Text).
		WillReturnRows(sqlmock.NewRows([]string{"id", "createdAt"}).AddRow(1, "2024-09-09T12:34:56Z"))

	comment, err := repo.CreateComment(context.Background(), input)

	require.NoError(t, err)

	expectedComment := &model.Comment{
		ID:        "1",
		PostID:    input.PostID,
		Sender:    &model.User{ID: input.SenderID},
		ReplyTo:   nil,
		Text:      input.Text,
		CreatedAt: "2024-09-09T12:34:56Z",
	}

	assert.Equal(t, expectedComment, comment)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestCreateComment_Error(t *testing.T) {
	// Мокируем базу данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresCommentRepository{Db: sqlxDB}

	input := model.NewComment{
		PostID:   "101",
		SenderID: "1",
		Text:     "This is a comment",
		ReplyTo:  nil,
	}

	mock.ExpectQuery(`^SELECT isCommentingAvailable FROM posts WHERE id = \$1$`).
		WithArgs(input.PostID).
		WillReturnError(sql.ErrConnDone)

	comment, err := repo.CreateComment(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, comment)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestComments_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresCommentRepository{Db: sqlxDB}

	limit := 10
	offset := 0

	rows := sqlmock.NewRows([]string{"id", "postid", "sender", "replyto", "text", "createdat", "id", "username"}).
		AddRow(1, 101, 1, nil, "This is a comment", "2024-09-09T12:34:56Z", 1, "user1").
		AddRow(2, 102, 2, nil, "Another comment", "2024-09-08T12:34:56Z", 2, "user2")

	mock.ExpectQuery("SELECT c.id, c.postid, c.sender, c.replyto, c.text, c.createdat, u.id, u.username").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	comments, err := repo.Comments(context.Background(), &limit, &offset)

	require.NoError(t, err)

	expectedComments := []*model.Comment{
		{
			ID:        "1",
			PostID:    "101",
			Sender:    &model.User{ID: "1", Username: "user1"},
			ReplyTo:   nil,
			Text:      "This is a comment",
			CreatedAt: "2024-09-09T12:34:56Z",
		},
		{
			ID:        "2",
			PostID:    "102",
			Sender:    &model.User{ID: "2", Username: "user2"},
			ReplyTo:   nil,
			Text:      "Another comment",
			CreatedAt: "2024-09-08T12:34:56Z",
		},
	}

	assert.ElementsMatch(t, expectedComments, comments)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestComments_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresCommentRepository{Db: sqlxDB}

	limit := 10
	offset := 0

	mock.ExpectQuery("SELECT c.id, c.postid, c.sender, c.replyto, c.text, c.createdat, u.id, u.username").
		WithArgs(limit, offset).
		WillReturnError(sql.ErrConnDone)

	comments, err := repo.Comments(context.Background(), &limit, &offset)

	require.Error(t, err)
	assert.Nil(t, comments)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
