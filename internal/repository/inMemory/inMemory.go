package inmemory

import (
	inmemoryComment "postus/internal/repository/inMemory/comment"
	inmemoryPost "postus/internal/repository/inMemory/post"
	inmemoryUser "postus/internal/repository/inMemory/user"
)

type Storage struct {
	Posts    *inmemoryPost.DataStore
	Comments *inmemoryComment.DataStore
	Users    *inmemoryUser.DataStore
}

func New() (*Storage, error) {
	usersStorage := inmemoryUser.New()
	return &Storage{
		Posts:    inmemoryPost.New(usersStorage),
		Comments: inmemoryComment.New(usersStorage),
		Users:    usersStorage,
	}, nil
}
