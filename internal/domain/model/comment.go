package model

import "time"

type Comment struct {
	ID              int64
	Body            string
	User            User
	PostID          int64
	PublicationTime time.Time
}

type Comments struct {
	Comments       []Comment
	HiddenComments bool
}
