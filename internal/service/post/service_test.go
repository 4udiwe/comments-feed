package post_service

import (
	"context"
	"errors"
	"testing"

	grapqlerrors "github.com/4udiwe/commnets-feed/internal/graph/errors"
	"github.com/4udiwe/commnets-feed/internal/graph/model"
	"github.com/4udiwe/commnets-feed/internal/repository"
	mock_post_service "github.com/4udiwe/commnets-feed/internal/service/post/mocks"
	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func assertGraphQLError(t *testing.T, err error, expectedMessage string, expectedCode string) {
	require.NotNil(t, err)
	gqlErr, ok := err.(*grapqlerrors.GraphQLError)
	require.True(t, ok, "expected GraphQLError, got %T", err)
	assert.Equal(t, expectedMessage, gqlErr.Message)
	assert.Equal(t, expectedCode, gqlErr.Code)
}

func TestCreatePost(t *testing.T) {
	tests := []struct {
		name           string
		input          model.CreatePostInput
		setupMocks     func(*gomock.Controller) PostRepository
		expectedErr    *struct{ Message, Code string }
		expectedResult *model.Post
		validate       func(*testing.T, *model.Post)
	}{
		{
			name: "успешное создание поста",
			input: model.CreatePostInput{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				return mockRepo
			},
			expectedErr: nil,
			validate: func(t *testing.T, post *model.Post) {
				assert.Equal(t, "Test Post", post.Title)
				assert.Equal(t, "This is a test post", post.Content)
				assert.True(t, post.CommentsEnabled)
				assert.NotEmpty(t, post.ID)
				assert.NotEmpty(t, post.CreatedAt)
			},
		},
		{
			name: "создание поста с явно отключенными комментариями",
			input: model.CreatePostInput{
				Title:           "Test Post",
				Content:         "This is a test post",
				CommentsEnabled: graphql.OmittableOf(boolPtr(false)),
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				return mockRepo
			},
			expectedErr: nil,
			validate: func(t *testing.T, post *model.Post) {
				assert.False(t, post.CommentsEnabled)
			},
		},
		{
			name: "создание поста с явно включенными комментариями",
			input: model.CreatePostInput{
				Title:           "Test Post",
				Content:         "This is a test post",
				CommentsEnabled: graphql.OmittableOf(boolPtr(true)),
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				return mockRepo
			},
			expectedErr: nil,
			validate: func(t *testing.T, post *model.Post) {
				assert.True(t, post.CommentsEnabled)
			},
		},
		{
			name: "пустой заголовок",
			input: model.CreatePostInput{
				Title:   "",
				Content: "This is a test post",
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "title cannot be empty", Code: "VALIDATION_ERROR"},
		},
		{
			name: "пустой контент",
			input: model.CreatePostInput{
				Title:   "Test Post",
				Content: "",
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "content cannot be empty", Code: "VALIDATION_ERROR"},
		},
		{
			name: "оба поля пусты",
			input: model.CreatePostInput{
				Title:   "",
				Content: "",
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "title cannot be empty", Code: "VALIDATION_ERROR"},
		},
		{
			name: "ошибка дублирования ключа",
			input: model.CreatePostInput{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(repository.ErrDuplicateKey)
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "post with this title already exists", Code: "CONFLICT"},
		},
		{
			name: "ошибка базы данных при создании",
			input: model.CreatePostInput{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("connection error"))
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "cannot create post", Code: "INTERNAL_SERVER_ERROR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.setupMocks(ctrl)
			service := NewPostService(repo)

			result, err := service.CreatePost(context.Background(), tt.input)

			if tt.expectedErr != nil {
				assertGraphQLError(t, err, tt.expectedErr.Message, tt.expectedErr.Code)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestGetPost(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		setupMocks     func(*gomock.Controller) PostRepository
		expectedErr    *struct{ Message, Code string }
		expectedResult *model.Post
	}{
		{
			name:   "успешное получение поста",
			postID: "post-1",
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				post := &model.Post{
					ID:              "post-1",
					Title:           "Test Post",
					Content:         "This is a test post",
					CommentsEnabled: true,
					CreatedAt:       "2026-04-18T12:00:00Z",
				}
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(post, nil)
				return mockRepo
			},
			expectedErr: nil,
			expectedResult: &model.Post{
				ID:              "post-1",
				Title:           "Test Post",
				Content:         "This is a test post",
				CommentsEnabled: true,
				CreatedAt:       "2026-04-18T12:00:00Z",
			},
		},
		{
			name:   "пост не найден",
			postID: "non-existent",
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "non-existent").
					Return(nil, repository.ErrPostNotFound)
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "post not found", Code: "NOT_FOUND"},
		},
		{
			name:   "ошибка базы данных",
			postID: "post-1",
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "post-1").
					Return(nil, errors.New("connection error"))
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "cannot fetch post", Code: "INTERNAL_SERVER_ERROR"},
		},
		{
			name:   "получение поста с отключенными комментариями",
			postID: "post-no-comments",
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				post := &model.Post{
					ID:              "post-no-comments",
					Title:           "No Comments Post",
					Content:         "Comments are disabled",
					CommentsEnabled: false,
					CreatedAt:       "2026-04-18T12:00:00Z",
				}
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "post-no-comments").
					Return(post, nil)
				return mockRepo
			},
			expectedErr: nil,
			expectedResult: &model.Post{
				ID:              "post-no-comments",
				Title:           "No Comments Post",
				Content:         "Comments are disabled",
				CommentsEnabled: false,
				CreatedAt:       "2026-04-18T12:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.setupMocks(ctrl)
			service := NewPostService(repo)

			result, err := service.GetPost(context.Background(), tt.postID)

			if tt.expectedErr != nil {
				assertGraphQLError(t, err, tt.expectedErr.Message, tt.expectedErr.Code)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.ID, result.ID)
				assert.Equal(t, tt.expectedResult.Title, result.Title)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.CommentsEnabled, result.CommentsEnabled)
			}
		})
	}
}

func TestListPosts(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		offset        int
		setupMocks    func(*gomock.Controller) PostRepository
		expectedErr   *struct{ Message, Code string }
		expectedCount int
	}{
		{
			name:   "успешное получение списка постов",
			limit:  10,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				posts := []*model.Post{
					{
						ID:              "post-1",
						Title:           "Post 1",
						Content:         "Content 1",
						CommentsEnabled: true,
						CreatedAt:       "2026-04-18T12:00:00Z",
					},
					{
						ID:              "post-2",
						Title:           "Post 2",
						Content:         "Content 2",
						CommentsEnabled: true,
						CreatedAt:       "2026-04-18T11:00:00Z",
					},
					{
						ID:              "post-3",
						Title:           "Post 3",
						Content:         "Content 3",
						CommentsEnabled: false,
						CreatedAt:       "2026-04-18T10:00:00Z",
					},
				}
				mockRepo.EXPECT().
					List(gomock.Any(), 10, 0).
					Return(posts, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 3,
		},
		{
			name:   "пустой список постов",
			limit:  10,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					List(gomock.Any(), 10, 0).
					Return([]*model.Post{}, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 0,
		},
		{
			name:   "получение постов со смещением",
			limit:  5,
			offset: 10,
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				posts := []*model.Post{
					{
						ID:              "post-11",
						Title:           "Post 11",
						Content:         "Content 11",
						CommentsEnabled: true,
						CreatedAt:       "2026-04-18T02:00:00Z",
					},
				}
				mockRepo.EXPECT().
					List(gomock.Any(), 5, 10).
					Return(posts, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 1,
		},
		{
			name:   "ошибка базы данных",
			limit:  10,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				mockRepo.EXPECT().
					List(gomock.Any(), 10, 0).
					Return(nil, errors.New("connection error"))
				return mockRepo
			},
			expectedErr: &struct{ Message, Code string }{Message: "cannot fetch posts", Code: "INTERNAL_SERVER_ERROR"},
		},
		{
			name:   "получение одного поста",
			limit:  1,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				posts := []*model.Post{
					{
						ID:              "post-1",
						Title:           "Post 1",
						Content:         "Content 1",
						CommentsEnabled: true,
						CreatedAt:       "2026-04-18T12:00:00Z",
					},
				}
				mockRepo.EXPECT().
					List(gomock.Any(), 1, 0).
					Return(posts, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 1,
		},
		{
			name:   "получение большого количества постов",
			limit:  100,
			offset: 0,
			setupMocks: func(ctrl *gomock.Controller) PostRepository {
				mockRepo := mock_post_service.NewMockPostRepository(ctrl)
				var posts []*model.Post
				for i := 1; i <= 100; i++ {
					posts = append(posts, &model.Post{
						ID:              "post-" + string(rune(i)),
						Title:           "Post " + string(rune(i)),
						Content:         "Content " + string(rune(i)),
						CommentsEnabled: true,
						CreatedAt:       "2026-04-18T12:00:00Z",
					})
				}
				mockRepo.EXPECT().
					List(gomock.Any(), 100, 0).
					Return(posts, nil)
				return mockRepo
			},
			expectedErr:   nil,
			expectedCount: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.setupMocks(ctrl)
			service := NewPostService(repo)

			result, err := service.ListPosts(context.Background(), tt.limit, tt.offset)

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

func boolPtr(b bool) *bool {
	return &b
}
