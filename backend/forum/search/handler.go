package search

import (
    "bufio"
    "database/sql"
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
)

type UserSearchResult struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Avatar   string `json:"avatar"`
}

type Service struct {
    DB *sql.DB
}

func loadNamedQuery(filename, name string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer f.Close()

    prefix := "-- name: " + name
    var lines []string
    var inBlock bool

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        trimmed := strings.TrimSpace(line)

        if strings.HasPrefix(trimmed, "-- name: ") {
            if inBlock {
                break
            }
            if trimmed == prefix {
                inBlock = true
            }
            continue
        }

        if inBlock {
            if strings.HasPrefix(trimmed, "--") || trimmed == "" {
                if len(lines) == 0 {
                    continue
                }
                break
            }
            lines = append(lines, line)
        }
    }

    if err := scanner.Err(); err != nil {
        return "", err
    }

    return strings.Join(lines, "\n"), nil
}

func (s *Service) SearchUsersAPI(c *gin.Context, currentUserID int) {
    q := strings.TrimSpace(c.Query("q"))
    if q == "" {
        c.JSON(http.StatusOK, []UserSearchResult{})
        return
    }

    query, err := loadNamedQuery("backend/database/schema.sql", "user_search_by_username")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "query load error"})
        return
    }

    pattern := "%" + q + "%"

    rows, err := s.DB.Query(query, currentUserID, pattern)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    defer rows.Close()

    results := []UserSearchResult{}
    for rows.Next() {
        var r UserSearchResult
        rows.Scan(&r.ID, &r.Username, &r.Avatar)
        results = append(results, r)
    }

    c.JSON(http.StatusOK, results)
}

func (s *Service) SearchPageHandler(c *gin.Context) {
    c.HTML(http.StatusOK, "search.html", gin.H{})
}