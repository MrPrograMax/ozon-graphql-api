package test

import (
	"context"
	"database/sql"
	"ozon-graphql-api/graph/model"
	"ozon-graphql-api/internal/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePost_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresPostRepository{Db: sqlxDB}

	input := model.NewPost{
		Title:                 "Test Title",
		Text:                  "Test Text",
		UserID:                "123",
		IsCommentingAvailable: new(bool),
	}

	rows := sqlmock.NewRows([]string{"id", "createdAt"}).
		AddRow(1, "2024-09-09T12:34:56Z")

	mock.ExpectQuery("INSERT INTO posts").
		WithArgs(input.Title, input.Text, input.UserID, input.IsCommentingAvailable).
		WillReturnRows(rows)

	post, err := repo.CreatePost(context.Background(), input)

	require.NoError(t, err)

	assert.Equal(t, "1", post.ID)
	assert.Equal(t, input.Title, post.Title)
	assert.Equal(t, input.Text, post.Text)
	assert.Equal(t, "2024-09-09T12:34:56Z", post.CreatedAt)
	assert.Equal(t, *input.IsCommentingAvailable, post.IsCommentingAvailable)
	assert.Equal(t, input.UserID, post.CreatedBy.ID)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestCreatePost_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresPostRepository{Db: sqlxDB}

	input := model.NewPost{
		Title:                 "Test Title",
		Text:                  "Test Text",
		UserID:                "123",
		IsCommentingAvailable: new(bool),
	}

	mock.ExpectQuery("INSERT INTO posts").
		WithArgs(input.Title, input.Text, input.UserID, input.IsCommentingAvailable).
		WillReturnError(sql.ErrConnDone)

	post, err := repo.CreatePost(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, post)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestPosts_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresPostRepository{Db: sqlxDB}

	limit := 10
	offset := 0

	rows := sqlmock.NewRows([]string{"id", "title", "text", "createdat", "iscommentingavailable", "userid", "username"}).
		AddRow(1, "Title1", "Text1", "2024-09-09T12:34:56Z", true, 123, "User1").
		AddRow(2, "Title2", "Text2", "2024-09-08T12:34:56Z", false, 124, "User2")

	mock.ExpectQuery("SELECT p.id, p.title, p.text, p.createdAt, p.isCommentingAvailable, u.id as userId, u.username").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	posts, err := repo.Posts(context.Background(), &limit, &offset)

	require.NoError(t, err)

	expectedPosts := []*model.Post{
		{
			ID:                    "1",
			Title:                 "Title1",
			Text:                  "Text1",
			CreatedAt:             "2024-09-09T12:34:56Z",
			IsCommentingAvailable: true,
			CreatedBy: &model.User{
				ID:       "123",
				Username: "User1",
			},
		},
		{
			ID:                    "2",
			Title:                 "Title2",
			Text:                  "Text2",
			CreatedAt:             "2024-09-08T12:34:56Z",
			IsCommentingAvailable: false,
			CreatedBy: &model.User{
				ID:       "124",
				Username: "User2",
			},
		},
	}

	assert.ElementsMatch(t, expectedPosts, posts)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestPosts_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	repo := &repository.PostgresPostRepository{Db: sqlxDB}

	limit := 10
	offset := 0

	mock.ExpectQuery("SELECT p.id, p.title, p.text, p.createdAt, p.isCommentingAvailable, u.id as userId, u.username").
		WithArgs(limit, offset).
		WillReturnError(sql.ErrConnDone)

	posts, err := repo.Posts(context.Background(), &limit, &offset)

	require.Error(t, err)
	assert.Nil(t, posts)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
