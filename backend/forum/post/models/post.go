package models

import "time"

type Mode = string

type Post struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	GameID    int        `json:"game_id"`
	Mode      string     `json:"mode"` // "Chill" ou "Ranked"
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	ImagePath *string    `json:"image_path,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Champs enrichis (jointures) — remplis par les requêtes de liste
	AuthorName   string     `json:"author_name"`
	GameName     string     `json:"game_name"`
	Categories   []Category `json:"categories"`
	LikeCount    int        `json:"like_count"`
	DislikeCount int        `json:"dislike_count"`
	CommentCount int        `json:"comment_count"`
	UserReaction string     `json:"user_reaction"` // "like", "dislike" ou ""
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Game struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	GameTypeID int    `json:"game_type_id"`
}
