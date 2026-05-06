package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var DB *sql.DB

func InitDB(db *sql.DB) {
	DB = db
}

func getUserIDFromSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, err
	}

	var userID int
	err = DB.QueryRow(`
		SELECT user_id
		FROM sessions
		WHERE session_id = ?
	`, cookie.Value).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func SteamLinkHandler(w http.ResponseWriter, r *http.Request) {
	_, err := getUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Vous devez être connecté", http.StatusUnauthorized)
		return
	}

	baseURL := "http://localhost:8080"
	returnTo := baseURL + "/auth/steam/callback"
	realm := baseURL

	params := url.Values{}
	params.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	params.Set("openid.mode", "checkid_setup")
	params.Set("openid.return_to", returnTo)
	params.Set("openid.realm", realm)
	params.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	params.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")

	redirectURL := "https://steamcommunity.com/openid/login?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func SteamCallbackHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Vous devez être connecté", http.StatusUnauthorized)
		return
	}

	err = validateSteamOpenID(r)
	if err != nil {
		http.Error(w, "Authentification Steam invalide", http.StatusUnauthorized)
		return
	}

	claimedID := r.URL.Query().Get("openid.claimed_id")
	steamID, err := extractSteamID(claimedID)
	if err != nil {
		http.Error(w, "SteamID invalide", http.StatusBadRequest)
		return
	}

	var existingUserID int
	err = DB.QueryRow("SELECT id FROM users WHERE steam_id = ?", steamID).Scan(&existingUserID)
	if err == nil && existingUserID != userID {
		http.Error(w, "Ce compte Steam est déjà lié à un autre utilisateur", http.StatusConflict)
		return
	}

	_, err = DB.Exec("UPDATE users SET steam_id = ? WHERE id = ?", steamID, userID)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func validateSteamOpenID(r *http.Request) error {
	values := url.Values{}

	for key, val := range r.URL.Query() {
		if strings.HasPrefix(key, "openid.") && len(val) > 0 {
			values.Set(key, val[0])
		}
	}

	values.Set("openid.mode", "check_authentication")

	resp, err := http.PostForm("https://steamcommunity.com/openid/login", values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "is_valid:true") {
		return fmt.Errorf("openid not valid")
	}

	return nil
}

func extractSteamID(claimedID string) (string, error) {
	re := regexp.MustCompile(`^https?://steamcommunity\.com/openid/id/(\d+)$`)
	matches := re.FindStringSubmatch(claimedID)
	if len(matches) != 2 {
		return "", fmt.Errorf("invalid claimed id")
	}
	return matches[1], nil
}