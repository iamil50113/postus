package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"postus/internal/domain/model"
	"postus/internal/repository"
)

func (s *Storage) User(ctx context.Context, uid int64) (*model.User, error) {
	const op = "repository.postgres.user.User"
	query := `SELECT name from users WHERE id = $1`

	var u model.User
	err := s.db.QueryRow(ctx, query, uid).Scan(&u.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrorUserNotFound
		}
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return &u, nil

	//rows, err := s.db.Query(ctx, query, uid)
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()
	//
	//var c *model.User
	//for rows.Next() {
	//	err := rows.Scan(&c.Name)
	//	if err != nil {
	//		return nil, err
	//	}
	//	c.ID = uid
	//}
	//
	//return c, nil
}