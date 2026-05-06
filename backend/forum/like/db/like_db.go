package db

import (
	"database/sql"
	"fmt"
)

type LikeDB struct {
	DB *sql.DB
}

// -------------------------------------------------------
// TOGGLE LIKE/DISLIKE SUR UN POST
// -------------------------------------------------------
// Logique :
//   - Si l'user n'a pas encore réagi → on insère la réaction
//   - Si l'user a déjà la même réaction → on la supprime (toggle off)
//   - Si l'user a la réaction opposée → on remplace

func (l *LikeDB) TogglePostReaction(postID, userID int, reactionType string) error {
	if reactionType != "like" && reactionType != "dislike" {
		return fmt.Errorf("reaction_type invalide : %s", reactionType)
	}

	var existing string
	err := l.DB.QueryRow(`
		SELECT reaction_type FROM post_reactions
		WHERE post_id = ? AND user_id = ?
	`, postID, userID).Scan(&existing)

	if err == sql.ErrNoRows {
		// Aucune réaction existante → on insère
		_, err = l.DB.Exec(`
			INSERT INTO post_reactions (post_id, user_id, reaction_type)
			VALUES (?, ?, ?)
		`, postID, userID, reactionType)
		return err
	}
	if err != nil {
		return fmt.Errorf("TogglePostReaction select: %w", err)
	}

	if existing == reactionType {
		// Même réaction → on supprime (toggle off)
		_, err = l.DB.Exec(`
			DELETE FROM post_reactions
			WHERE post_id = ? AND user_id = ?
		`, postID, userID)
		return err
	}

	// Réaction opposée → on met à jour
	_, err = l.DB.Exec(`
		UPDATE post_reactions
		SET reaction_type = ?, created_at = CURRENT_TIMESTAMP
		WHERE post_id = ? AND user_id = ?
	`, reactionType, postID, userID)
	return err
}

// -------------------------------------------------------
// TOGGLE LIKE/DISLIKE SUR UN COMMENTAIRE
// -------------------------------------------------------

func (l *LikeDB) ToggleCommentReaction(commentID, userID int, reactionType string) error {
	if reactionType != "like" && reactionType != "dislike" {
		return fmt.Errorf("reaction_type invalide : %s", reactionType)
	}

	var existing string
	err := l.DB.QueryRow(`
		SELECT reaction_type FROM comment_reactions
		WHERE comment_id = ? AND user_id = ?
	`, commentID, userID).Scan(&existing)

	if err == sql.ErrNoRows {
		_, err = l.DB.Exec(`
			INSERT INTO comment_reactions (comment_id, user_id, reaction_type)
			VALUES (?, ?, ?)
		`, commentID, userID, reactionType)
		return err
	}
	if err != nil {
		return fmt.Errorf("ToggleCommentReaction select: %w", err)
	}

	if existing == reactionType {
		_, err = l.DB.Exec(`
			DELETE FROM comment_reactions
			WHERE comment_id = ? AND user_id = ?
		`, commentID, userID)
		return err
	}

	_, err = l.DB.Exec(`
		UPDATE comment_reactions
		SET reaction_type = ?, created_at = CURRENT_TIMESTAMP
		WHERE comment_id = ? AND user_id = ?
	`, reactionType, commentID, userID)
	return err
}

// -------------------------------------------------------
// COMPTER LES RÉACTIONS D'UN POST (utile pour la réponse JSON)
// -------------------------------------------------------

type ReactionCount struct {
	Likes        int    `json:"likes"`
	Dislikes     int    `json:"dislikes"`
	UserReaction string `json:"user_reaction"` // "like", "dislike" ou ""
}

func (l *LikeDB) GetPostReactions(postID, userID int) (ReactionCount, error) {
	var rc ReactionCount

	err := l.DB.QueryRow(`
		SELECT
			COUNT(DISTINCT CASE WHEN reaction_type = 'like'    THEN user_id END),
			COUNT(DISTINCT CASE WHEN reaction_type = 'dislike' THEN user_id END),
			COALESCE((
				SELECT reaction_type FROM post_reactions
				WHERE post_id = ? AND user_id = ?
			), '')
		FROM post_reactions
		WHERE post_id = ?
	`, postID, userID, postID).Scan(&rc.Likes, &rc.Dislikes, &rc.UserReaction)

	if err != nil {
		return rc, fmt.Errorf("GetPostReactions: %w", err)
	}
	return rc, nil
}

func (l *LikeDB) GetCommentReactions(commentID, userID int) (ReactionCount, error) {
	var rc ReactionCount

	err := l.DB.QueryRow(`
		SELECT
			COUNT(DISTINCT CASE WHEN reaction_type = 'like'    THEN user_id END),
			COUNT(DISTINCT CASE WHEN reaction_type = 'dislike' THEN user_id END),
			COALESCE((
				SELECT reaction_type FROM comment_reactions
				WHERE comment_id = ? AND user_id = ?
			), '')
		FROM comment_reactions
		WHERE comment_id = ?
	`, commentID, userID, commentID).Scan(&rc.Likes, &rc.Dislikes, &rc.UserReaction)

	if err != nil {
		return rc, fmt.Errorf("GetCommentReactions: %w", err)
	}
	return rc, nil
}
