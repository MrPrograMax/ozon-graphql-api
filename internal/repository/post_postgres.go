package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"ozon-graphql-api/graph/model"
	"strconv"
)

type PostgresPostRepository struct {
	Db *sqlx.DB
}

func NewPostgresPostRepo(db *sqlx.DB) *PostgresPostRepository {
	return &PostgresPostRepository{
		Db: db,
	}
}

func (r *PostgresPostRepository) Posts(ctx context.Context, limit, offset *int) ([]*model.Post, error) {
	queryLimit := *limit
	queryOffset := *offset

	postFields := `p.id, p.title, p.text, p.createdAt, p.isCommentingAvailable, u.id as userId, u.username`

	query := fmt.Sprintf(`SELECT %s FROM %s p JOIN %s u ON p.createdBy = u.id 
                              ORDER BY p.createdAt DESC LIMIT $1 OFFSET $2`,
		postFields, postsTable, usersTable)

	// Промежуточная структура для маппинга
	var dbPosts []dbPostStruct

	err := r.Db.SelectContext(ctx, &dbPosts, query, queryLimit, queryOffset)
	if err != nil {
		return nil, err
	}

	var posts []*model.Post

	for _, dbPost := range dbPosts {
		//Заполняем данные о посте
		postId := strconv.Itoa(dbPost.ID)
		post := &model.Post{
			ID:                    postId,
			Title:                 dbPost.Title,
			Text:                  dbPost.Text,
			CreatedAt:             dbPost.CreatedAt,
			IsCommentingAvailable: dbPost.IsCommentingAvailable,
		}

		//Дописываем в модель информацию о пользователе
		if dbPost.UserID != nil && dbPost.Username != nil {
			userId := strconv.Itoa(*dbPost.UserID)
			post.CreatedBy = &model.User{
				ID:       userId,
				Username: *dbPost.Username,
			}
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostgresPostRepository) PostByID(ctx context.Context, id int) (*model.Post, error) {
	postFields := `p.id, p.title, p.text, p.createdAt, p.isCommentingAvailable, u.id as userId, u.username`
	postQuery := fmt.Sprintf(`SELECT %s FROM %s p JOIN %s u ON p.createdBy = u.id WHERE p.id = $1`,
		postFields, postsTable, usersTable)

	var dbPost dbPostStruct
	if err := r.Db.GetContext(ctx, &dbPost, postQuery, id); err != nil {
		return nil, fmt.Errorf("error fetching post: %w", err)
	}

	// Рекурсивный обход комментариев для конкретного поста
	fields := `c.id, c.text, c.replyTo, c.sender, c.createdAt`
	name := `comment_tree`
	ctFields := `ct.id, ct.text, ct.replyto, ct.sender, ct.createdat, u.id as userid, u.username`
	commentQuery := fmt.Sprintf(`
    WITH RECURSIVE %s AS (
        SELECT %s
        FROM %s c
        WHERE c.postId = $1 AND c.replyTo IS NULL
        UNION ALL
        SELECT %s
        FROM %s c
        INNER JOIN %s ct ON c.replyTo = ct.id
    )
    SELECT %s FROM %s ct
    JOIN %s u ON ct.sender = u.id
    ORDER BY createdAt;`,
		name, fields, commentsTable, fields, commentsTable, name,
		ctFields, name, usersTable)

	// Промежуточная структура для маппинга
	var dbComments []dbCommentStruct
	if err := r.Db.SelectContext(ctx, &dbComments, commentQuery, id); err != nil {
		return nil, err
	}

	// Преобразование в иерархическую структуру
	commentMap := make(map[int]*model.Comment)
	var rootComments []*model.Comment

	for _, c := range dbComments {
		comment := &model.Comment{
			ID:        strconv.Itoa(c.ID),
			PostID:    strconv.Itoa(c.PostID),
			Text:      c.Text,
			ReplyTo:   nil,
			CreatedAt: c.CreatedAt,
			Sender: &model.User{
				ID:       strconv.Itoa(c.SenderID),
				Username: *c.Username,
			},
			Replies: []*model.Comment{},
		}
		commentMap[c.ID] = comment
	}

	for _, c := range dbComments {
		comment := commentMap[c.ID]
		if c.ReplyTo != nil {
			parent, exists := commentMap[*c.ReplyTo]
			if exists {
				comment.ReplyTo = parent
				parent.Replies = append(parent.Replies, comment)
			}
		} else {
			rootComments = append(rootComments, comment)
		}
	}

	postId := strconv.Itoa(dbPost.ID)
	post := &model.Post{
		ID:                    postId,
		Title:                 dbPost.Title,
		Text:                  dbPost.Text,
		CreatedAt:             dbPost.CreatedAt,
		IsCommentingAvailable: dbPost.IsCommentingAvailable,
		CreatedBy: &model.User{
			ID:       strconv.Itoa(*dbPost.UserID),
			Username: *dbPost.Username,
		},
		Comments: rootComments,
	}

	return post, nil
}

func (r *PostgresPostRepository) CreatePost(ctx context.Context, input model.NewPost) (*model.Post, error) {
	postFields := `title, text, createdBy, isCommentingAvailable`

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES ($1, $2, $3, $4) RETURNING id, createdAt`,
		postsTable, postFields)

	var postId int
	var createdAt string

	err := r.Db.QueryRowContext(ctx, query,
		input.Title, input.Text, input.UserID, input.IsCommentingAvailable).Scan(&postId, &createdAt)
	if err != nil {
		return nil, err
	}

	id := strconv.Itoa(postId)
	post := &model.Post{
		ID:                    id,
		Title:                 input.Title,
		Text:                  input.Text,
		CreatedAt:             createdAt,
		IsCommentingAvailable: *input.IsCommentingAvailable,
		CreatedBy: &model.User{
			ID: input.UserID,
		},
	}

	return post, nil
}
