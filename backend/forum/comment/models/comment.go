package models

import "time"

type Comment struct {
	ID        int        `json:"id"`
	PostID    int        `json:"post_id"`
	UserID    int        `json:"user_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Champs enrichis
	AuthorName   string `json:"author_name"`
	LikeCount    int    `json:"like_count"`
	DislikeCount int    `json:"dislike_count"`
	UserReaction string `json:"user_reaction"`
}
