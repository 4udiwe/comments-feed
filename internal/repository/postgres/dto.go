package postgres_repository

import (
	"time"

	"github.com/4udiwe/comments-feed/internal/graph/model"
)

// PostDTO представляет пост как он хранится в базе данных
type PostDTO struct {
	ID              string    `db:"id"`
	Title           string    `db:"title"`
	Content         string    `db:"content"`
	CommentsEnabled bool      `db:"comments_enabled"`
	CreatedAt       time.Time `db:"created_at"`
}

// ToEntity преобразует PostDTO в модель Post
func (dto *PostDTO) ToEntity() *model.Post {
	if dto == nil {
		return nil
	}

	return &model.Post{
		ID:              dto.ID,
		Title:           dto.Title,
		Content:         dto.Content,
		CommentsEnabled: dto.CommentsEnabled,
		CreatedAt:       dto.CreatedAt.String(),
		Comments:        []*model.Comment{},
	}
}

// CommentDTO представляет комментарий как он хранится в базе данных
type CommentDTO struct {
	ID        string    `db:"id"`
	PostID    string    `db:"post_id"`
	ParentID  *string   `db:"parent_id"`
	Text      string    `db:"text"`
	CreatedAt time.Time `db:"created_at"`
}

// ToEntity преобразует CommentDTO в модель Comment
func (dto *CommentDTO) ToEntity() *model.Comment {
	if dto == nil {
		return nil
	}

	return &model.Comment{
		ID:        dto.ID,
		PostID:    dto.PostID,
		ParentID:  dto.ParentID,
		Text:      dto.Text,
		CreatedAt: dto.CreatedAt.String(),
		Children:  []*model.Comment{},
	}
}
