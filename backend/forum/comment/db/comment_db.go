package db

import (
	"database/sql"
	"fmt"
	"socialauth/backend/forum/comment/models"
)

type CommentDB struct {
	DB *sql.DB
}

// -------------------------------------------------------
// LISTE DES COMMENTAIRES D'UN POST
// -------------------------------------------------------

func (c *CommentDB) GetCommentsByPost(postID, viewerUserID int) ([]models.Comment, error) {
	query := `
		SELECT
			cm.id, cm.post_id, cm.user_id, cm.content, cm.created_at, cm.updated_at,
			u.username AS author_name,
			COUNT(DISTINCT CASE WHEN cr.reaction_type = 'like'    THEN cr.user_id END) AS like_count,
			COUNT(DISTINCT CASE WHEN cr.reaction_type = 'dislike' THEN cr.user_id END) AS dislike_count,
			COALESCE(my_r.reaction_type, '') AS user_reaction
		FROM comments cm
		JOIN users u ON u.id = cm.user_id
		LEFT JOIN comment_reactions cr ON cr.comment_id = cm.id
		LEFT JOIN comment_reactions my_r ON my_r.comment_id = cm.id AND my_r.user_id = ?
		WHERE cm.post_id = ?
		GROUP BY cm.id
		ORDER BY cm.created_at ASC
	`
	rows, err := c.DB.Query(query, viewerUserID, postID)
	if err != nil {
		return nil, fmt.Errorf("GetCommentsByPost: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var cm models.Comment
		err := rows.Scan(
			&cm.ID, &cm.PostID, &cm.UserID, &cm.Content,
			&cm.CreatedAt, &cm.UpdatedAt,
			&cm.AuthorName,
			&cm.LikeCount, &cm.DislikeCount,
			&cm.UserReaction,
		)
		if err != nil {
			return nil, fmt.Errorf("GetCommentsByPost scan: %w", err)
		}
		comments = append(comments, cm)
	}
	return comments, nil
}

// -------------------------------------------------------
// CRÉER UN COMMENTAIRE
// -------------------------------------------------------

func (c *CommentDB) CreateComment(postID, userID int, content string) (int64, error) {
	res, err := c.DB.Exec(`
		INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)
	`, postID, userID, content)
	if err != nil {
		return 0, fmt.Errorf("CreateComment: %w", err)
	}
	return res.LastInsertId()
}

// -------------------------------------------------------
// ÉDITER UN COMMENTAIRE (proprio uniquement)
// -------------------------------------------------------

func (c *CommentDB) UpdateComment(commentID, userID int, content string) error {
	res, err := c.DB.Exec(`
		UPDATE comments
		SET content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, content, commentID, userID)
	if err != nil {
		return fmt.Errorf("UpdateComment: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("not found or forbidden")
	}
	return nil
}

// -------------------------------------------------------
// SUPPRIMER UN COMMENTAIRE (proprio uniquement)
// -------------------------------------------------------

func (c *CommentDB) DeleteComment(commentID, userID int) error {
	res, err := c.DB.Exec(`
		DELETE FROM comments WHERE id = ? AND user_id = ?
	`, commentID, userID)
	if err != nil {
		return fmt.Errorf("DeleteComment: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("not found or forbidden")
	}
	return nil
}

// -------------------------------------------------------
// VÉRIFIER QU'UN COMMENTAIRE APPARTIENT BIEN À UN POST
// (sécurité : évite de modifier un commentaire d'un autre post)
// -------------------------------------------------------

func (c *CommentDB) CommentBelongsToPost(commentID, postID int) (bool, error) {
	var count int
	err := c.DB.QueryRow(`
		SELECT COUNT(*) FROM comments WHERE id = ? AND post_id = ?
	`, commentID, postID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
