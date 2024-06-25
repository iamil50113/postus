package inmemorymodel

import "time"

type Post struct {
	Title             string
	Body              string
	UserID            int64
	PublicationTime   time.Time
	CommentPermission bool
}
