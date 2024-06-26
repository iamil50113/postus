package inmemory

import (
	inmemoryComment "postus/internal/repository/inMemory/comment"
	inmemoryPost "postus/internal/repository/inMemory/post"
	inmemoryUser "postus/internal/repository/inMemory/user"
)

type Storage struct {
	posts    *inmemoryPost.DataStore
	comments *inmemoryComment.DataStore
	users    *inmemoryUser.DataStore
}

func New(commentsPaginationLimit int) (*Storage, error) {
	usersStorage := inmemoryUser.New()
	return &Storage{
		posts:    inmemoryPost.New(usersStorage),
		comments: inmemoryComment.New(usersStorage, commentsPaginationLimit),
		users:    usersStorage,
	}, nil
}
