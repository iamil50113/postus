package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Storage, error) {
	return &Storage{db: pool}, nil
}

func createTables(db *pgxpool.Pool) {
	db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS post(
		id BIGSERIAL PRIMARY KEY,
	    title TEXT,
		body TEXT,
		user_id BIGSERIAL,
		publication_time TIMESTAMP,
		comment_permission BOOLEAN
	                               );

	CREATE TABLE IF NOT EXISTS comment(
		id BIGSERIAL PRIMARY KEY,
		body TEXT,
		user_id BIGSERIAL,
		post_id BIGSERIAL,
		parent_comment_id BIGSERIAL NULL,
		publication_time TIMESTAMP
	                                   );

	CREATE TABLE IF NOT EXISTS users(
		id BIGSERIAL PRIMARY KEY,
		name TEXT
	                                  );
`)
}
