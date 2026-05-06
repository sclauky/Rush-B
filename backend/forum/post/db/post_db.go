package db

import (
	"database/sql"
	"fmt"
	"socialauth/backend/forum/post/models"
)

type PostDB struct {
	DB *sql.DB
}

// -------------------------------------------------------
// LISTE DES POSTS (avec filtres optionnels)
// -------------------------------------------------------

type PostFilter struct {
	CategoryID *int // filtrer par catégorie
	UserID     *int // filtrer par auteur (posts créés par moi)
	LikedBy    *int // filtrer par posts likés par moi
	GameID     *int // filtrer par jeu
}

func (p *PostDB) GetPosts(filter PostFilter, viewerUserID int) ([]models.Post, error) {
	// On construit la requête dynamiquement selon les filtres
	query := `
		SELECT DISTINCT
			p.id, p.user_id, p.game_id, p.mode, p.title, p.content,
			p.image_path, p.created_at, p.updated_at,
			u.username AS author_name,
			g.name AS game_name,
			COUNT(DISTINCT CASE WHEN pr.reaction_type = 'like'    THEN pr.user_id END) AS like_count,
			COUNT(DISTINCT CASE WHEN pr.reaction_type = 'dislike' THEN pr.user_id END) AS dislike_count,
			COUNT(DISTINCT c.id) AS comment_count,
			COALESCE(my_r.reaction_type, '') AS user_reaction
		FROM posts p
		JOIN users u ON u.id = p.user_id
		JOIN games g ON g.id = p.game_id
		LEFT JOIN post_categories pc ON pc.post_id = p.id
		LEFT JOIN post_reactions pr ON pr.post_id = p.id
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN post_reactions my_r ON my_r.post_id = p.id AND my_r.user_id = ?
	`
	args := []any{viewerUserID}

	// Jointure supplémentaire si on filtre par liked
	if filter.LikedBy != nil {
		query += ` JOIN post_reactions liked_r ON liked_r.post_id = p.id AND liked_r.reaction_type = 'like' AND liked_r.user_id = ?`
		args = append(args, *filter.LikedBy)
	}

	query += ` WHERE 1=1`

	if filter.CategoryID != nil {
		query += ` AND pc.category_id = ?`
		args = append(args, *filter.CategoryID)
	}
	if filter.UserID != nil {
		query += ` AND p.user_id = ?`
		args = append(args, *filter.UserID)
	}
	if filter.GameID != nil {
		query += ` AND p.game_id = ?`
		args = append(args, *filter.GameID)
	}

	query += `
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`

	rows, err := p.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetPosts query: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID, &post.UserID, &post.GameID, &post.Mode,
			&post.Title, &post.Content, &post.ImagePath,
			&post.CreatedAt, &post.UpdatedAt,
			&post.AuthorName, &post.GameName,
			&post.LikeCount, &post.DislikeCount, &post.CommentCount,
			&post.UserReaction,
		)
		if err != nil {
			return nil, fmt.Errorf("GetPosts scan: %w", err)
		}
		// On récupère les catégories de chaque post
		post.Categories, err = p.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// -------------------------------------------------------
// UN SEUL POST PAR ID
// -------------------------------------------------------

func (p *PostDB) GetPostByID(postID, viewerUserID int) (*models.Post, error) {
	query := `
		SELECT
			p.id, p.user_id, p.game_id, p.mode, p.title, p.content,
			p.image_path, p.created_at, p.updated_at,
			u.username AS author_name,
			g.name AS game_name,
			COUNT(DISTINCT CASE WHEN pr.reaction_type = 'like'    THEN pr.user_id END) AS like_count,
			COUNT(DISTINCT CASE WHEN pr.reaction_type = 'dislike' THEN pr.user_id END) AS dislike_count,
			COUNT(DISTINCT c.id) AS comment_count,
			COALESCE(my_r.reaction_type, '') AS user_reaction
		FROM posts p
		JOIN users u ON u.id = p.user_id
		JOIN games g ON g.id = p.game_id
		LEFT JOIN post_reactions pr ON pr.post_id = p.id
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN post_reactions my_r ON my_r.post_id = p.id AND my_r.user_id = ?
		WHERE p.id = ?
		GROUP BY p.id
	`
	var post models.Post
	err := p.DB.QueryRow(query, viewerUserID, postID).Scan(
		&post.ID, &post.UserID, &post.GameID, &post.Mode,
		&post.Title, &post.Content, &post.ImagePath,
		&post.CreatedAt, &post.UpdatedAt,
		&post.AuthorName, &post.GameName,
		&post.LikeCount, &post.DislikeCount, &post.CommentCount,
		&post.UserReaction,
	)
	if err == sql.ErrNoRows {
		return nil, nil // post inexistant → handler renverra 404
	}
	if err != nil {
		return nil, fmt.Errorf("GetPostByID: %w", err)
	}
	post.Categories, err = p.getCategoriesForPost(post.ID)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// -------------------------------------------------------
// CRÉATION D'UN POST
// -------------------------------------------------------

type CreatePostInput struct {
	UserID      int
	GameID      int
	Mode        string // "Chill" ou "Ranked"
	Title       string
	Content     string
	ImagePath   *string
	CategoryIDs []int
}

func (p *PostDB) CreatePost(input CreatePostInput) (int64, error) {
	tx, err := p.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("CreatePost begin tx: %w", err)
	}
	defer tx.Rollback() // annulé si on n'appelle pas Commit()

	res, err := tx.Exec(`
		INSERT INTO posts (user_id, game_id, mode, title, content, image_path)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.UserID, input.GameID, input.Mode, input.Title, input.Content, input.ImagePath)
	if err != nil {
		return 0, fmt.Errorf("CreatePost insert: %w", err)
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Associer les catégories
	for _, catID := range input.CategoryIDs {
		_, err := tx.Exec(`
			INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)
		`, postID, catID)
		if err != nil {
			return 0, fmt.Errorf("CreatePost category link: %w", err)
		}
	}

	return postID, tx.Commit()
}

// -------------------------------------------------------
// ÉDITION D'UN POST
// -------------------------------------------------------

type UpdatePostInput struct {
	PostID      int
	UserID      int // pour vérifier que c'est bien le proprio
	Mode        string
	Title       string
	Content     string
	ImagePath   *string
	CategoryIDs []int
}

func (p *PostDB) UpdatePost(input UpdatePostInput) error {
	tx, err := p.DB.Begin()
	if err != nil {
		return fmt.Errorf("UpdatePost begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE posts
		SET mode = ?, title = ?, content = ?, image_path = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, input.Mode, input.Title, input.Content, input.ImagePath, input.PostID, input.UserID)
	if err != nil {
		return fmt.Errorf("UpdatePost exec: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Soit le post n'existe pas, soit l'user n'est pas le proprio
		return fmt.Errorf("not found or forbidden")
	}

	// Remplacer les catégories
	_, err = tx.Exec(`DELETE FROM post_categories WHERE post_id = ?`, input.PostID)
	if err != nil {
		return fmt.Errorf("UpdatePost delete categories: %w", err)
	}
	for _, catID := range input.CategoryIDs {
		_, err := tx.Exec(`
			INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)
		`, input.PostID, catID)
		if err != nil {
			return fmt.Errorf("UpdatePost insert category: %w", err)
		}
	}

	return tx.Commit()
}

// -------------------------------------------------------
// SUPPRESSION D'UN POST
// -------------------------------------------------------

func (p *PostDB) DeletePost(postID, userID int) error {
	// ON DELETE CASCADE dans la BDD s'occupe des catégories, réactions, commentaires
	res, err := p.DB.Exec(`
		DELETE FROM posts WHERE id = ? AND user_id = ?
	`, postID, userID)
	if err != nil {
		return fmt.Errorf("DeletePost: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("not found or forbidden")
	}
	return nil
}

// -------------------------------------------------------
// UTILITAIRE INTERNE — catégories d'un post
// -------------------------------------------------------

func (p *PostDB) getCategoriesForPost(postID int) ([]models.Category, error) {
	rows, err := p.DB.Query(`
		SELECT c.id, c.name
		FROM categories c
		JOIN post_categories pc ON pc.category_id = c.id
		WHERE pc.post_id = ?
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("getCategoriesForPost: %w", err)
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// -------------------------------------------------------
// RÉCUPÉRER TOUS LES JEUX (pour le formulaire de création)
// -------------------------------------------------------

func (p *PostDB) GetAllGames() ([]models.Game, error) {
	rows, err := p.DB.Query(`SELECT id, name, game_type_id FROM games ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("GetAllGames: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		if err := rows.Scan(&g.ID, &g.Name, &g.GameTypeID); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

// -------------------------------------------------------
// RÉCUPÉRER TOUTES LES CATÉGORIES (pour le formulaire)
// -------------------------------------------------------

func (p *PostDB) GetAllCategories() ([]models.Category, error) {
	rows, err := p.DB.Query(`SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("GetAllCategories: %w", err)
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}
