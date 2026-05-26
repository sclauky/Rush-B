package swipe

import (
    "database/sql"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
)

type User struct {
    ID         int      `json:"id"`
    FirstName  string   `json:"first_name"`
    LastName   string   `json:"last_name"`
    Username   string   `json:"username"`
    Biography  string   `json:"biography"`
    GameTypes  []string `json:"game_types"`
    GamesPlayed []string `json:"games_played"`
    SkillLevel string   `json:"skill_level"`
    PlayStyle  string   `json:"play_style"`
    Photos     []string `json:"photos"`
}

type SwipeRequest struct {
    TargetID  int    `json:"target_id"`
    Direction string `json:"direction"`
}

type MatchResponse struct {
    Matched bool `json:"matched"`
    MatchID int  `json:"match_id,omitempty"`
}

type MatchEntry struct {
    User      User      `json:"user"`
    MatchID   int       `json:"match_id"`
    MatchedAt time.Time `json:"matched_at"`
}

type Service struct {
    DB *sql.DB
}

func NewService(db *sql.DB) *Service {
    return &Service{DB: db}
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
    r.GET("/profiles", s.getProfiles)
    r.POST("/action", s.handleSwipe)
    r.GET("/matches", s.getMatches)
    r.DELETE("/matches/:id", s.deleteMatch)
    r.POST("/photos", s.uploadPhoto)
    r.DELETE("/photos/:id", s.deletePhoto)
}

func (s *Service) getProfiles(c *gin.Context) {
    userID := currentUserID(c)

    rows, err := s.DB.Query(`
        SELECT u.id, u.first_name, u.last_name, u.username, COALESCE(u.biography, ''),
               COALESCE(up.skill_level, ''), COALESCE(up.play_style, '')
        FROM users u
        LEFT JOIN user_preferences up ON up.user_id = u.id
        WHERE u.id != ?
          AND u.id NOT IN (SELECT target_id FROM swipes WHERE swiper_id = ?)
        ORDER BY RANDOM()
        LIMIT 20
    `, userID, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    defer rows.Close()

    profiles := []User{}
    for rows.Next() {
        var u User
        rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.Biography, &u.SkillLevel, &u.PlayStyle)
        u.GameTypes = queryStrings(s.DB, `SELECT game_type FROM user_game_types WHERE user_id = ?`, u.ID)
        u.GamesPlayed = queryStrings(s.DB, `SELECT game_name FROM user_games WHERE user_id = ?`, u.ID)
        u.Photos = queryStrings(s.DB, `SELECT photo_path FROM user_photos WHERE user_id = ? ORDER BY position ASC LIMIT 6`, u.ID)
        profiles = append(profiles, u)
    }
    c.JSON(http.StatusOK, profiles)
}

func (s *Service) handleSwipe(c *gin.Context) {
    userID := currentUserID(c)

    var req SwipeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
        return
    }

    _, err := s.DB.Exec(
        `INSERT OR IGNORE INTO swipes (swiper_id, target_id, direction, created_at) VALUES (?, ?, ?, ?)`,
        userID, req.TargetID, req.Direction, time.Now(),
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }

    resp := MatchResponse{}
    if req.Direction == "like" {
        var reverseID int
        err := s.DB.QueryRow(
            `SELECT id FROM swipes WHERE swiper_id = ? AND target_id = ? AND direction = 'like'`,
            req.TargetID, userID,
        ).Scan(&reverseID)
        if err == nil {
            var matchID int
            err2 := s.DB.QueryRow(
                `SELECT id FROM matches WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)`,
                userID, req.TargetID, req.TargetID, userID,
            ).Scan(&matchID)
            if err2 == sql.ErrNoRows {
                res, _ := s.DB.Exec(
                    `INSERT INTO matches (user1_id, user2_id, created_at) VALUES (?, ?, ?)`,
                    userID, req.TargetID, time.Now(),
                )
                id64, _ := res.LastInsertId()
                matchID = int(id64)
            }
            resp.Matched = true
            resp.MatchID = matchID
        }
    }
    c.JSON(http.StatusOK, resp)
}

func (s *Service) getMatches(c *gin.Context) {
    userID := currentUserID(c)

    rows, err := s.DB.Query(`
        SELECT u.id, u.first_name, u.last_name, u.username, COALESCE(u.biography, ''),
               COALESCE(up.skill_level, ''), COALESCE(up.play_style, ''), m.id, m.created_at
        FROM matches m
        JOIN users u ON (u.id = CASE WHEN m.user1_id = ? THEN m.user2_id ELSE m.user1_id END)
        LEFT JOIN user_preferences up ON up.user_id = u.id
        WHERE m.user1_id = ? OR m.user2_id = ?
        ORDER BY m.created_at DESC
    `, userID, userID, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    defer rows.Close()

    matches := []MatchEntry{}
    for rows.Next() {
        var u User
        var me MatchEntry
        rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.Biography,
            &u.SkillLevel, &u.PlayStyle, &me.MatchID, &me.MatchedAt)
        u.GameTypes = queryStrings(s.DB, `SELECT game_type FROM user_game_types WHERE user_id = ?`, u.ID)
        u.Photos = queryStrings(s.DB, `SELECT photo_path FROM user_photos WHERE user_id = ? ORDER BY position ASC LIMIT 1`, u.ID)
        me.User = u
        matches = append(matches, me)
    }
    c.JSON(http.StatusOK, matches)
}

func (s *Service) deleteMatch(c *gin.Context) {
    userID := currentUserID(c)
    matchID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    var u1, u2 int
    err = s.DB.QueryRow(`SELECT user1_id, user2_id FROM matches WHERE id = ?`, matchID).Scan(&u1, &u2)
    if err != nil || (u1 != userID && u2 != userID) {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }

    s.DB.Exec(`DELETE FROM matches WHERE id = ?`, matchID)
    c.JSON(http.StatusOK, gin.H{"ok": "unmatched"})
}

func (s *Service) uploadPhoto(c *gin.Context) {
    userID := currentUserID(c)

    var count int
    s.DB.QueryRow(`SELECT COUNT(*) FROM user_photos WHERE user_id = ?`, userID).Scan(&count)
    if count >= 6 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "max 6 photos"})
        return
    }

    file, header, err := c.Request.FormFile("photo")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "no file"})
        return
    }
    defer file.Close()

    uploadDir := "frontend/web/uploads/swipe"
    os.MkdirAll(uploadDir, 0755)

    filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), header.Filename)
    dst, err := os.Create(filepath.Join(uploadDir, filename))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
        return
    }
    defer dst.Close()

    buf := make([]byte, 10<<20)
    n, _ := file.Read(buf)
    dst.Write(buf[:n])

    photoPath := "/uploads/swipe/" + filename
    s.DB.Exec(`INSERT INTO user_photos (user_id, photo_path, position) VALUES (?, ?, ?)`,
        userID, photoPath, count)

    c.JSON(http.StatusOK, gin.H{"path": photoPath})
}

func (s *Service) deletePhoto(c *gin.Context) {
    userID := currentUserID(c)
    photoID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    s.DB.Exec(`DELETE FROM user_photos WHERE id = ? AND user_id = ?`, photoID, userID)
    c.JSON(http.StatusOK, gin.H{"ok": "deleted"})
}

// ICI, à la fin du fichier:

func currentUserID(c *gin.Context) int {
    v, exists := c.Get("userID")
    if !exists {
        return 0
    }

    userID, ok := v.(int)
    if !ok {
        return 0
    }

    return userID
}

func queryStrings(db *sql.DB, query string, args ...interface{}) []string {
    rows, err := db.Query(query, args...)
    if err != nil {
        return []string{}
    }
    defer rows.Close()
    var out []string
    for rows.Next() {
        var s string
        rows.Scan(&s)
        out = append(out, s)
    }
    if out == nil {
        return []string{}
    }
    return out
}