package comment_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/4udiwe/commnets-feed/internal/broker"
	grapqlerrors "github.com/4udiwe/commnets-feed/internal/graph/errors"
	"github.com/4udiwe/commnets-feed/internal/graph/model"
	mock_transactor "github.com/4udiwe/commnets-feed/internal/mocks"
	mock_comment_service "github.com/4udiwe/commnets-feed/internal/service/comment/mocks"
	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func stringPtr(s string) *string {
	return &s
}

func assertGraphQLError(t *testing.T, err error, expectedMessage string, expectedCode string) {
	require.NotNil(t, err)
	gqlErr, ok := err.(*grapqlerrors.GraphQLError)
	require.True(t, ok, "expected GraphQLError, got %T", err)
	assert.Equal(t, expectedMessage, gqlErr.Message)
	assert.Equal(t, expectedCode, gqlErr.Code)
}

func TestCreateComment(t *testing.T) {
	tests := []struct {
		name           string
		input          model.CreateCommentInput
		setupMocks     func(*gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor)
		expectedErr    *struct{ Message, Code string }
		expectedResult *model.Comment
	}{
		{
			name: "успешное создание комментария",
			input: model.CreateCommentInput{
				PostID: "post-1",
				Text:   "Test comment",
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(&model.Post{ID: "post-1", CommentsEnabled: true}, nil)

				commentRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				broker.EXPECT().
					Publish(gomock.Any(), gomock.Any())

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: nil,
		},
		{
			name: "пустой текст комментария",
			input: model.CreateCommentInput{
				PostID: "post-1",
				Text:   "",
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "text cannot be empty", Code: "VALIDATION_ERROR"},
		},
		{
			name: "текст превышает максимальную длину",
			input: model.CreateCommentInput{
				PostID: "post-1",
				Text:   string(make([]byte, 2001)),
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "text must not exceed 2000 characters", Code: "VALIDATION_ERROR"},
		},
		{
			name: "пост не найден",
			input: model.CreateCommentInput{
				PostID: "non-existent",
				Text:   "Test comment",
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "non-existent").
					Return(nil, errors.New("not found"))

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "post not found", Code: "NOT_FOUND"},
		},
		{
			name: "комментарии отключены для поста",
			input: model.CreateCommentInput{
				PostID: "post-1",
				Text:   "Test comment",
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(&model.Post{ID: "post-1", CommentsEnabled: false}, nil)

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "comments are disabled for this post", Code: "FORBIDDEN"},
		},
		{
			name: "создание комментария с родителем",
			input: model.CreateCommentInput{
				PostID:   "post-1",
				Text:     "Reply comment",
				ParentID: graphql.OmittableOf(stringPtr("parent-comment-1")),
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(&model.Post{ID: "post-1", CommentsEnabled: true}, nil)

				commentRepo.EXPECT().
					GetByID(gomock.Any(), "parent-comment-1").
					Return(&model.Comment{ID: "parent-comment-1", PostID: "post-1"}, nil)

				commentRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				broker.EXPECT().
					Publish(gomock.Any(), gomock.Any())

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: nil,
		},
		{
			name: "родительский комментарий не найден",
			input: model.CreateCommentInput{
				PostID:   "post-1",
				Text:     "Reply comment",
				ParentID: graphql.OmittableOf(stringPtr("non-existent-parent")),
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(&model.Post{ID: "post-1", CommentsEnabled: true}, nil)

				commentRepo.EXPECT().
					GetByID(gomock.Any(), "non-existent-parent").
					Return(nil, errors.New("not found"))

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "parent comment not found", Code: "NOT_FOUND"},
		},
		{
			name: "родительский комментарий из другого поста",
			input: model.CreateCommentInput{
				PostID:   "post-1",
				Text:     "Reply comment",
				ParentID: graphql.OmittableOf(stringPtr("parent-comment-2")),
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(&model.Post{ID: "post-1", CommentsEnabled: true}, nil)

				commentRepo.EXPECT().
					GetByID(gomock.Any(), "parent-comment-2").
					Return(&model.Comment{ID: "parent-comment-2", PostID: "post-2"}, nil)

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "parent comment belongs to another post", Code: "BAD_REQUEST"},
		},
		{
			name: "ошибка при создании комментария",
			input: model.CreateCommentInput{
				PostID: "post-1",
				Text:   "Test comment",
			},
			setupMocks: func(ctrl *gomock.Controller) (CommentRepository, CommentBroker, PostRepository, *mock_transactor.MockTransactor) {
				commentRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				broker := mock_comment_service.NewMockCommentBroker(ctrl)
				postRepo := mock_comment_service.NewMockPostRepository(ctrl)
				txManager := mock_transactor.NewMockTransactor(ctrl)

				postRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(&model.Post{ID: "post-1", CommentsEnabled: true}, nil)

				commentRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))

				txManager.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return commentRepo, broker, postRepo, txManager
			},
			expectedErr: &struct{ Message, Code string }{Message: "cannot create comment", Code: "INTERNAL_SERVER_ERROR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentRepo, commentBroker, postRepo, txManager := tt.setupMocks(ctrl)
			service := NewCommentService(commentRepo, commentBroker, postRepo, txManager)

			result, err := service.CreateComment(context.Background(), tt.input)

			if tt.expectedErr != nil {
				assertGraphQLError(t, err, tt.expectedErr.Message, tt.expectedErr.Code)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "post-1", result.PostID)
				assert.Equal(t, tt.input.Text, result.Text)
			}
		})
	}
}

func TestGetCommentsByPost(t *testing.T) {
	tests := []struct {
		name          string
		postID        string
		limit         int
		offset        int
		setupMocks    func(*gomock.Controller) CommentRepository
		expectedErr   *struct{ Message, Code string }
		expectedCount int
	}{
		{
			name:   "успешное получение комментариев",
			postID: "post-1",
			limit:  10,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				comments := []*model.Comment{
					{ID: "comment-1", PostID: "post-1", Text: "Comment 1"},
					{ID: "comment-2", PostID: "post-1", Text: "Comment 2"},
				}
				mockRepo.EXPECT().
					GetByPostID(gomock.Any(), "post-1", 10, 0).
					Return(comments, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 2,
		},
		{
			name:   "пост не содержит комментариев",
			postID: "post-empty",
			limit:  10,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				mockRepo.EXPECT().
					GetByPostID(gomock.Any(), "post-empty", 10, 0).
					Return([]*model.Comment{}, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 0,
		},
		{
			name:   "ошибка при получении комментариев",
			postID: "post-1",
			limit:  10,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				mockRepo.EXPECT().
					GetByPostID(gomock.Any(), "post-1", 10, 0).
					Return(nil, errors.New("db error"))
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "cannot fetch comments", Code: "INTERNAL_SERVER_ERROR"},
		},
		{
			name:   "пагинация со смещением",
			postID: "post-1",
			limit:  5,
			offset: 10,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				mockRepo.EXPECT().
					GetByPostID(gomock.Any(), "post-1", 5, 10).
					Return([]*model.Comment{}, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentRepo := tt.setupMocks(ctrl)
			service := NewCommentService(commentRepo, nil, nil, nil)

			result, err := service.GetCommentsByPost(context.Background(), tt.postID, tt.limit, tt.offset)

			if tt.expectedErr != nil {
				assertGraphQLError(t, err, tt.expectedErr.Message, tt.expectedErr.Code)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestGetChildren(t *testing.T) {
	tests := []struct {
		name          string
		parentID      string
		limit         int
		offset        int
		setupMocks    func(*gomock.Controller) CommentRepository
		expectedErr   *struct{ Message, Code string }
		expectedCount int
	}{
		{
			name:     "успешное получение дочерних комментариев",
			parentID: "parent-1",
			limit:    10,
			offset:   0,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				children := []*model.Comment{
					{ID: "child-1", ParentID: stringPtr("parent-1"), Text: "Reply 1"},
					{ID: "child-2", ParentID: stringPtr("parent-1"), Text: "Reply 2"},
					{ID: "child-3", ParentID: stringPtr("parent-1"), Text: "Reply 3"},
				}
				mockRepo.EXPECT().
					GetChildren(gomock.Any(), "parent-1", 10, 0).
					Return(children, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 3,
		},
		{
			name:     "комментарий без ответов",
			parentID: "parent-no-replies",
			limit:    10,
			offset:   0,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				mockRepo.EXPECT().
					GetChildren(gomock.Any(), "parent-no-replies", 10, 0).
					Return([]*model.Comment{}, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 0,
		},
		{
			name:     "ошибка при получении дочерних комментариев",
			parentID: "parent-1",
			limit:    10,
			offset:   0,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				mockRepo.EXPECT().
					GetChildren(gomock.Any(), "parent-1", 10, 0).
					Return(nil, errors.New("db error"))
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "cannot fetch comments", Code: "INTERNAL_SERVER_ERROR"},
		},
		{
			name:     "пагинация дочерних комментариев",
			parentID: "parent-1",
			limit:    2,
			offset:   1,
			setupMocks: func(ctrl *gomock.Controller) CommentRepository {
				mockRepo := mock_comment_service.NewMockCommentRepository(ctrl)
				children := []*model.Comment{
					{ID: "child-2", ParentID: stringPtr("parent-1"), Text: "Reply 2"},
					{ID: "child-3", ParentID: stringPtr("parent-1"), Text: "Reply 3"},
				}
				mockRepo.EXPECT().
					GetChildren(gomock.Any(), "parent-1", 2, 1).
					Return(children, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentRepo := tt.setupMocks(ctrl)
			service := NewCommentService(commentRepo, nil, nil, nil)

			result, err := service.GetChildren(context.Background(), tt.parentID, tt.limit, tt.offset)

			if tt.expectedErr != nil {
				assertGraphQLError(t, err, tt.expectedErr.Message, tt.expectedErr.Code)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestSubscribeToComments(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		setupMocks  func(*gomock.Controller) (CommentBroker, func())
		expectedErr error
		testFn      func(*testing.T, <-chan *model.Comment)
	}{
		{
			name:   "успешная подписка на комментарии",
			postID: "post-1",
			setupMocks: func(ctrl *gomock.Controller) (CommentBroker, func()) {
				mockBroker := mock_comment_service.NewMockCommentBroker(ctrl)
				ch := make(chan *broker.Comment, 1)
				unsubscribe := func() {
					close(ch)
				}
				mockBroker.EXPECT().
					Subscribe("post-1").
					Return(ch, unsubscribe)

				return mockBroker, func() {
				}
			},
			expectedErr: nil,
			testFn: func(t *testing.T, out <-chan *model.Comment) {
				assert.NotNil(t, out)
			},
		},
		{
			name:   "контекст отмены останавливает подписку",
			postID: "post-1",
			setupMocks: func(ctrl *gomock.Controller) (CommentBroker, func()) {
				mockBroker := mock_comment_service.NewMockCommentBroker(ctrl)
				ch := make(chan *broker.Comment, 1)
				unsubscribe := func() {
					close(ch)
				}
				mockBroker.EXPECT().
					Subscribe("post-1").
					Return(ch, unsubscribe)

				return mockBroker, func() {}
			},
			expectedErr: nil,
			testFn: func(t *testing.T, out <-chan *model.Comment) {
				assert.NotNil(t, out)
			},
		},
		{
			name:   "получение сообщения через подписку",
			postID: "post-1",
			setupMocks: func(ctrl *gomock.Controller) (CommentBroker, func()) {
				mockBroker := mock_comment_service.NewMockCommentBroker(ctrl)
				ch := make(chan *broker.Comment, 1)
				unsubscribe := func() {
					close(ch)
				}
				mockBroker.EXPECT().
					Subscribe("post-1").
					Return(ch, unsubscribe)

				// Функция для отправки сообщения в канал
				sendMessage := func() {
					ch <- &broker.Comment{
						ID:      "comment-1",
						Content: "Test message",
						PostID:  "post-1",
					}
				}

				return mockBroker, sendMessage
			},
			expectedErr: nil,
			testFn: func(t *testing.T, out <-chan *model.Comment) {
				// Ожидаем получить сообщение из выходного канала
				select {
				case msg := <-out:
					assert.NotNil(t, msg)
					assert.Equal(t, "comment-1", msg.ID)
					assert.Equal(t, "Test message", msg.Text)
					assert.Equal(t, "post-1", msg.PostID)
				case <-time.After(time.Second):
					t.Fatal("timeout waiting for message")
				}
			},
		},
		{
			name:   "канал закрывается при отмене контекста",
			postID: "post-1",
			setupMocks: func(ctrl *gomock.Controller) (CommentBroker, func()) {
				mockBroker := mock_comment_service.NewMockCommentBroker(ctrl)
				ch := make(chan *broker.Comment, 1)
				unsubscribe := func() {
					close(ch)
				}
				mockBroker.EXPECT().
					Subscribe("post-1").
					Return(ch, unsubscribe)

				return mockBroker, func() {}
			},
			expectedErr: nil,
			testFn: func(t *testing.T, out <-chan *model.Comment) {
				// Проверяем что канал закрывается
				_, ok := <-out
				assert.False(t, ok, "expected channel to be closed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentBroker, testFn := tt.setupMocks(ctrl)
			service := NewCommentService(nil, commentBroker, nil, nil)

			ctx := context.Background()

			// Используем отмену контекста
			if tt.name == "канал закрывается при отмене контекста" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				defer cancel()

				result, err := service.SubscribeToComments(ctx, tt.postID)

				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)

					// Отменяем контекст
					cancel()

					// Даем время на выполнение горутины
					time.Sleep(100 * time.Millisecond)

					// Проверяем что канал закрыт
					_, ok := <-result
					assert.False(t, ok, "expected channel to be closed after context cancellation")
				}
			} else {
				result, err := service.SubscribeToComments(ctx, tt.postID)

				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					require.NotNil(t, result)

					if tt.name == "получение сообщения через подписку" {
						// Отправляем сообщение
						testFn()
					}

					// Выполняем специфичный для теста код
					tt.testFn(t, result)
				}
			}
		})
	}
}
