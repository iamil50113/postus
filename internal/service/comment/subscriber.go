package comment

import (
	"context"
	"postus/internal/domain/model"
	"sync"
)

//type Ch struct {
//	channel  chan *model.Comment
//	isClosed *bool
//}

type PostCommentsSubs struct {
	sync.RWMutex
	subs map[*chan *model.Comment]struct{}
}

type Subscriber struct {
	sync.RWMutex
	posts        map[int64]*PostCommentsSubs
	postProvider PostProvider
}

func newSubscriber(postProvider PostProvider) *Subscriber {
	return &Subscriber{
		postProvider: postProvider,
	}
}

func (s *Subscriber) NewCommentAlert(comment *model.Comment) {
	go func() {
		s.RLock()

		if p, ok := s.posts[comment.PostID]; ok {
			s.posts[comment.PostID].RLock()

			for k, _ := range p.subs {
				*k <- comment
			}

			s.posts[comment.PostID].RUnlock()
		}

		s.RUnlock()
	}()

}

func (s *Subscriber) NewSubscribe(ctx context.Context, postID int64) (<-chan *model.Comment, error) {
	newCommentEvent := make(chan *model.Comment, 1)

	s.Lock()

	p, ok := s.posts[postID]
	if ok {
		p.Lock()

		p.subs[&newCommentEvent] = struct{}{}

		p.Unlock()

	} else {

		subs := make(map[*chan *model.Comment]struct{})
		subs[&newCommentEvent] = struct{}{}

		s.posts[postID] = &PostCommentsSubs{subs: subs}
	}
	s.Unlock()

	go func() {
		<-ctx.Done()

		s.RLock()

		s.posts[postID].Lock()

		delete(s.posts[postID].subs, &newCommentEvent)

		s.posts[postID].Unlock()

		s.RUnlock()

		close(newCommentEvent)
	}()

	return newCommentEvent, nil
}
