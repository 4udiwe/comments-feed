-- +goose Up
CREATE TABLE post (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(512) NOT NULL,
    content TEXT NOT NULL,
    comments_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_post_created_at ON post(created_at DESC);


CREATE TABLE comment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comment(id) ON DELETE CASCADE,
    text TEXT NOT NULL CHECK (LENGTH(text) <= 2000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comment_post_id ON comment(post_id);
CREATE INDEX idx_comment_parent_id ON comment(parent_id);
CREATE INDEX idx_comment_post_id_parent_id ON comment(post_id, parent_id);
CREATE INDEX idx_comment_created_at ON comment(created_at DESC);
CREATE INDEX idx_comment_post_created ON comment(post_id, created_at DESC);
-- +goose Down
DROP TABLE post;
DROP TABLE comment;
