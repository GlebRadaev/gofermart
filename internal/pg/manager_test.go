package pg

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*MockDatabase, *MockTX, *Manager) {
	ctrl := gomock.NewController(t)
	mockDB := NewMockDatabase(ctrl)
	mockTx := NewMockTX(ctrl)
	manager := NewTXManager(mockDB)
	defer ctrl.Finish()
	return mockDB, mockTx, manager
}

func TestTXManager_Begin(t *testing.T) {
	mockDB, mockTx, manager := NewMock(t)

	testCases := []struct {
		name           string
		mockSetup      func()
		ctxSetup       func() context.Context
		fn             func(ctx context.Context) error
		expectedErrMsg string
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockDB.EXPECT().Begin(gomock.Any()).Return(mockTx, nil).Times(1)
				mockTx.EXPECT().Commit(gomock.Any()).Return(nil).Times(1)
			},
			ctxSetup: func() context.Context {
				return context.Background()
			},
			fn:             func(ctx context.Context) error { return nil },
			expectedErrMsg: "",
		},
		{
			name: "With active transaction",
			mockSetup: func() {
				mockDB.EXPECT().Begin(gomock.Any()).Times(0)
			},
			ctxSetup: func() context.Context {
				ctx := context.Background()
				return With(ctx, WithTransaction(&MockTX{}))
			},
			fn:             func(ctx context.Context) error { return nil },
			expectedErrMsg: "",
		},
		{
			name: "Error on Begin",
			mockSetup: func() {
				mockDB.EXPECT().Begin(gomock.Any()).Return(nil, errors.New("failed to begin transaction")).Times(1)
			},
			ctxSetup: func() context.Context {
				return context.Background()
			},
			fn:             func(ctx context.Context) error { return nil },
			expectedErrMsg: "pg: can't begin tx",
		},
		{
			name: "Error in transaction function",
			mockSetup: func() {
				mockDB.EXPECT().Begin(gomock.Any()).Return(mockTx, nil).Times(1)
				mockTx.EXPECT().Rollback(gomock.Any()).Return(nil).Times(1)
			},
			ctxSetup: func() context.Context {
				return context.Background()
			},
			fn:             func(ctx context.Context) error { return errors.New("some error") },
			expectedErrMsg: "some error",
		},
		{
			name: "Error on Commit",
			mockSetup: func() {
				mockDB.EXPECT().Begin(gomock.Any()).Return(mockTx, nil).Times(1)
				mockTx.EXPECT().Commit(gomock.Any()).Return(errors.New("commit error")).Times(1)
				mockTx.EXPECT().Rollback(gomock.Any()).Return(nil).Times(1)
			},
			ctxSetup: func() context.Context {
				return context.Background()
			},
			fn:             func(ctx context.Context) error { return nil },
			expectedErrMsg: "pg: can't commit tx: commit error",
		},
		{
			name: "Error on Rollback",
			mockSetup: func() {
				mockDB.EXPECT().Begin(gomock.Any()).Return(mockTx, nil).Times(1)
				mockTx.EXPECT().Rollback(gomock.Any()).Return(errors.New("simulated rollback error")).Times(1)
			},
			ctxSetup: func() context.Context {
				return context.Background()
			},
			fn:             func(ctx context.Context) error { return errors.New("simulated function error") },
			expectedErrMsg: "pg: can't rollback tx: simulated rollback error, original error: simulated function error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			ctx := tc.ctxSetup()

			err := manager.Begin(ctx, tc.fn)

			if tc.expectedErrMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			}
		})
	}
}
