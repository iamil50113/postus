package inmemory

import (
	inmemorymodel "postus/internal/repository/inMemory/model"
	"sync"
)

type Storage struct {
	posts           []*inmemorymodel.Post
	comments        []*inmemorymodel.Comment
	users           []string
	paginationLimit int
	muPosts         sync.Mutex
	muComments      sync.Mutex
	muUsers         sync.Mutex
}

func New() (*Storage, error) {
	users := make([]string, 1, 100)
	users = append(users, "ivan")
	users = append(users, "egor")
	users = append(users, "alex")

	return &Storage{posts: make([]*inmemorymodel.Post, 1, 100),
		comments: make([]*inmemorymodel.Comment, 1, 100),
		users:    users}, nil
}
