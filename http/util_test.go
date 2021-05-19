package http

import (
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/golang/mock/gomock"
)

// newTestEnv calls services.NewTestEnv and mocks notification's Broadcast and Notify.
func newTestEnv(t *testing.T) (*services.Env, func()) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	n := mocks.NewMockNotifier(mock)
	n.EXPECT().Broadcast(any, any, any, any, any, any, any, any).AnyTimes()
	n.EXPECT().Notify(any, any).AnyTimes()
	e.Notifier = n
	return e, func() { mock.Finish() }
}
