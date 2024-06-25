package comment

import (
	"context"
	"postus/internal/domain/model"
	"sync"
	"time"
)

type Ch struct {
	channel  chan *model.Comment
	isClosed *bool
}

type Subscriber struct {
	subs map[int64][]Ch
	sync.Mutex
}

func (s *Subscriber) NewPostAlert(comment *model.Comment) {
	go func() {
		s.Mutex.Lock()
		for i := 0; i < len(s.subs[comment.PostID]); i++ {
			if !*s.subs[comment.PostID][i].isClosed {
				s.subs[comment.PostID][i].channel <- comment
			}
		}
		s.Mutex.Unlock()
	}()

}

func (s *Subscriber) NewSubscribe(ctx context.Context, postID int64) (<-chan *model.Comment, error) {
	newCommentEvent := make(chan *model.Comment, 1)
	closed := false
	go func() {
		// Handle deregistration of the channel here. Note the `defer`
		defer func() {
			close(newCommentEvent)
			closed = true
		}()

		for {
			time.Sleep(1 * time.Second)

			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	s.Mutex.Lock()

	if len(s.subs[postID]) == 0 {
		s.subs[postID] = make([]Ch, 0, 1)
	}

	s.subs[postID] = append(s.subs[postID], Ch{channel: newCommentEvent, isClosed: &closed})

	s.Mutex.Unlock()

	return newCommentEvent, nil
}
