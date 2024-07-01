package comment_test

import (
	"github.com/golang/mock/gomock"
	gomocks "postus/internal/service/comment/mocks"
	"testing"
)

func testNewComment(t *testing.T) {
	type mockBehavior func(comProvider *gomocks.MockCommentProvider, comSaver *gomocks.MockCommentSaver, usProvider *gomocks.MockUserProvider, postProvider *gomocks.MockPostProvider, commentID int64, err error)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockComProvider := gomocks.NewMockCommentProvider(mockCtrl)
	mockPostProvider := gomocks.NewMockPostProvider(mockCtrl)
	userProvider := gomocks.NewMockUserProvider(mockCtrl)
	mockComSaver := gomocks.NewMockCommentSaver(mockCtrl)

	testTable := []struct {
		name         string
		newCommentID int64
		err          error
		mockBehavior mockBehavior
	}{{name: "OK",
		newCommentID: 1,
		err:          nil,
	},
	}
}
