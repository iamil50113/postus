package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"postus/internal/domain/model"
	"postus/internal/repository"
	"time"
)

func (s *Storage) ChildExist(ctx context.Context, commentID int64) (bool, error) {
	const op = "repository.postgres.comment.ChildExist"
	query := `SELECT EXISTS (SELECT * FROM comment WHERE parent_comment_id = $1);`

	var exists bool
	err := s.db.QueryRow(ctx, query, commentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return exists, nil
}

func (s *Storage) ChildCommentsForParentCommentIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	const op = "repository.postgres.comment.GetChildCommentsForParentCommentIDWithCursor"

	query := `SELECT comment.id, body, post_id, publication_time, user_id, name
				FROM comment
    			JOIN users ON comment.user_id = users.id
				WHERE comment.id > $1 AND parent_comment_id = $2
				LIMIT $3`
	rows, err := s.db.Query(ctx, query, cursor, id, limit+1)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	cms := make([]model.Comment, 0, 6)
	for rows.Next() {
		c := model.Comment{}
		err := rows.Scan(&c.ID, &c.Body, &c.PostID, &c.PublicationTime, &c.User.ID, &c.User.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		cms = append(cms, c)
	}

	if len(cms) > limit {
		return &model.Comments{Comments: cms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: cms, HiddenComments: false}, nil
}

func (s *Storage) CommentsForPostIDWithCursor(ctx context.Context, id int64, cursor int64, limit int) (*model.Comments, error) {
	const op = "repository.postgres.comment.GetCommentsForPostIDWithCursor"

	query := `SELECT comment.id, body, publication_time, user_id, name
				FROM comment
    			JOIN users ON comment.user_id = users.id
                WHERE comment.id > $1 AND post_id = $2 AND parent_comment_id IS NULL
                LIMIT $3`
	rows, err := s.db.Query(ctx, query, cursor, id, limit+1)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	cms := make([]model.Comment, 0, 6)
	for rows.Next() {
		c := model.Comment{}
		err := rows.Scan(&c.ID, &c.Body, &c.PublicationTime, &c.User.ID, &c.User.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		cms = append(cms, c)
	}
	if len(cms) > limit {
		return &model.Comments{Comments: cms[:limit], HiddenComments: true}, nil
	}

	return &model.Comments{Comments: cms, HiddenComments: false}, nil
}

func (s *Storage) NewComment(ctx context.Context, uid int64, postID int64, body string, publicationTime time.Time) (int64, error) {
	const op = "repository.postgres.comment.NewComment"

	query := `INSERT INTO comment (body, user_id, post_id, publication_time)
VALUES (@body, @userID, @postID, @publicationTime) RETURNING id`

	args := pgx.NamedArgs{
		"body":            body,
		"userID":          uid,
		"postID":          postID,
		"publicationTime": publicationTime,
	}

	var id int64
	err := s.db.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return id, nil
}

func (s *Storage) NewChildComment(ctx context.Context, uid int64, postID int64, body string, parentCommentID int64, publicationTime time.Time) (int64, error) {
	const op = "repository.postgres.comment.NewComment"

	query := `INSERT INTO comment (body, user_id, post_id, publication_time, parent_comment_id)
VALUES (@body, @userID, @postID, @publicationTime, @parentCommentID) RETURNING id`

	args := pgx.NamedArgs{
		"body":            body,
		"userID":          uid,
		"postID":          postID,
		"publicationTime": publicationTime,
		"parentCommentID": parentCommentID,
	}

	var id int64
	err := s.db.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return id, nil
}

func (s *Storage) Comment(ctx context.Context, id int64) (*model.Comment, error) {
	const op = "repository.postgres.comment.Comment"
	query := `SELECT body, user_id, post_id, publication_time from comment WHERE id = $1`
	var c model.Comment
	err := s.db.QueryRow(ctx, query, id).Scan(&c.Body, &c.User.ID, &c.PostID, &c.PublicationTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrorCommentNotFound
		}
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return &c, nil
}
