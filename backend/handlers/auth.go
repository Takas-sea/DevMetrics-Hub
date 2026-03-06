package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Takas-sea/DevMetrics-Hub/models"
	"github.com/Takas-sea/DevMetrics-Hub/utils"
)

type AuthHandler struct {
	DB *sql.DB
}

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" || strings.Contains(strings.ToLower(clientID), "test") {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "GitHub OAuth is not configured. Set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET in .env",
		})
		return
	}

	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://localhost:3000/auth/callback"
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("scope", "read:user")

	authURL := "https://github.com/login/oauth/authorize?" + values.Encode()

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	accessToken, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	githubUser, err := h.fetchGitHubUser(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := h.createOrUpdateUser(githubUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionID := uuid.New().String()
	_, err = h.DB.Exec(
		"INSERT INTO sessions (id, user_id, token, expires_at) VALUES ($1, $2, $3, $4)",
		sessionID, user.ID, token, time.Now().Add(24*time.Hour),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) exchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("github oauth credentials are not configured")
	}

	accessTokenURL := fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		clientID, clientSecret, code,
	)

	req, _ := http.NewRequest("POST", accessTokenURL, nil)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (h *AuthHandler) fetchGitHubUser(accessToken string) (*GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *AuthHandler) createOrUpdateUser(githubUser *GitHubUser) (*models.User, error) {
	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, github_id, username, email, avatar_url, bio, created_at, updated_at FROM users WHERE github_id = $1",
		githubUser.ID,
	).Scan(&user.ID, &user.GitHubID, &user.Username, &user.Email, &user.AvatarURL, &user.Bio, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		user.ID = uuid.New().String()
		user.GitHubID = githubUser.ID
		user.Username = githubUser.Login
		user.Email = githubUser.Email
		user.AvatarURL = githubUser.AvatarURL
		user.Bio = githubUser.Bio
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		_, err = h.DB.Exec(
			"INSERT INTO users (id, github_id, username, email, avatar_url, bio, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			user.ID, user.GitHubID, user.Username, user.Email, user.AvatarURL, user.Bio, user.CreatedAt, user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "logout success",
	})
}
