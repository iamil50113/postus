package model

import "time"

type Post struct {
	ID                int64
	Title             string
	Body              string
	User              User
	PublicationTime   time.Time
	CommentPermission bool
}
