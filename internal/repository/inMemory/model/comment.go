package inmemorymodel

import "time"

type Comment struct {
	Body            string
	UserID          int64
	PostID          int64
	ParentCommentID int64
	PublicationTime time.Time
}
