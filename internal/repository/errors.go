package repository

import "errors"

var (
	ErrorPostNotFound    = errors.New("post not found")
	ErrorCommentNotFound = errors.New("comment not found")
	ErrorUserNotFound    = errors.New("user not found")
)
