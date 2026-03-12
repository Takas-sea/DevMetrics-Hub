package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Takas-sea/DevMetrics-Hub/db"
	"github.com/Takas-sea/DevMetrics-Hub/utils"
)

type ActivityHandler struct {
	cacheMu sync.RWMutex
	cache   map[string]cacheEntry
	ttl     time.Duration
	db      *sql.DB
}

type cacheEntry struct {
	value     *activityResponse
	expiresAt time.Time
}

func NewActivityHandler(dbConn *sql.DB) *ActivityHandler {
	cacheTTL := 5 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("ACTIVITY_CACHE_TTL_SECONDS")); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err == nil && seconds >= 0 {
			cacheTTL = time.Duration(seconds) * time.Second
		}
	}

	return &ActivityHandler{
		cache: make(map[string]cacheEntry),
		ttl:   cacheTTL,
		db:    dbConn,
	}
}

type githubEvent struct {
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
	Payload struct {
		Commits []struct{} `json:"commits"`
	} `json:"payload"`
}

type dailyActivityPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type repositoryActivity struct {
	Name    string `json:"name"`
	Commits int    `json:"commits"`
}

type activitySummary struct {
	TotalContributions int `json:"total_contributions"`
	AverageDaily       int `json:"average_daily"`
}

type activityResponse struct {
	Daily              []dailyActivityPoint `json:"daily"`
	Summary            activitySummary      `json:"summary"`
	RecentRepositories []repositoryActivity `json:"recent_repositories"`
}

func (h *ActivityHandler) GetMyActivities(c *gin.Context) {
	tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	days := 5
	if raw := c.Query("days"); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 1 || parsed > 30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "days must be between 1 and 30"})
			return
		}
		days = parsed
	}

	cacheKey := fmt.Sprintf("%s:%d", strings.ToLower(strings.TrimSpace(claims.Username)), days)
	if cached, ok := h.getCachedActivity(cacheKey); ok {
		c.Header("X-Activity-Cache", "HIT")
		c.JSON(http.StatusOK, cached)
		return
	}

	res, err := h.fetchRecentActivities(claims.Username, days)
	if err != nil {
		fallback := h.getFallbackActivities(claims.UserID, days)
		if fallback != nil {
			c.Header("X-Activity-Source", "DB-FALLBACK")
			c.JSON(http.StatusOK, fallback)
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	h.persistActivities(claims.UserID, claims.Username, res)
	h.setCachedActivity(cacheKey, res)
	c.Header("X-Activity-Cache", "MISS")

	c.JSON(http.StatusOK, res)
}

func (h *ActivityHandler) getCachedActivity(key string) (*activityResponse, bool) {
	if h == nil || h.ttl <= 0 {
		return nil, false
	}

	h.cacheMu.RLock()
	entry, ok := h.cache[key]
	h.cacheMu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		h.cacheMu.Lock()
		delete(h.cache, key)
		h.cacheMu.Unlock()
		return nil, false
	}

	return entry.value, true
}

func (h *ActivityHandler) setCachedActivity(key string, value *activityResponse) {
	if h == nil || h.ttl <= 0 {
		return
	}

	h.cacheMu.Lock()
	if h.cache == nil {
		h.cache = make(map[string]cacheEntry)
	}
	h.cache[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(h.ttl),
	}
	h.cacheMu.Unlock()
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("authorization header must be in format: Bearer <token>")
	}

	return strings.TrimSpace(parts[1]), nil
}

func (h *ActivityHandler) fetchRecentActivities(username string, days int) (*activityResponse, error) {
	since := time.Now().UTC().AddDate(0, 0, -(days - 1))
	dailyCounts := map[string]int{}
	for i := 0; i < days; i++ {
		dateKey := since.AddDate(0, 0, i).Format("2006-01-02")
		dailyCounts[dateKey] = 0
	}

	repoCounts := map[string]int{}
	client := &http.Client{Timeout: 10 * time.Second}

	for page := 1; page <= 3; page++ {
		endpoint := fmt.Sprintf("https://api.github.com/users/%s/events/public?per_page=100&page=%d", url.PathEscape(username), page)
		req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "DevMetrics-Hub")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch github events: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("github events API returned status %d", resp.StatusCode)
		}

		var events []githubEvent
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode github events: %w", err)
		}
		resp.Body.Close()

		if len(events) == 0 {
			break
		}

		for _, event := range events {
			if event.Type != "PushEvent" {
				continue
			}

			eventTime, err := time.Parse(time.RFC3339, event.CreatedAt)
			if err != nil {
				continue
			}

			eventTime = eventTime.UTC()
			if eventTime.Before(since) {
				continue
			}

			count := len(event.Payload.Commits)
			if count == 0 {
				count = 1
			}

			dateKey := eventTime.Format("2006-01-02")
			if _, exists := dailyCounts[dateKey]; exists {
				dailyCounts[dateKey] += count
			}

			repoName := strings.TrimSpace(event.Repo.Name)
			if repoName != "" {
				repoCounts[repoName] += count
			}
		}
	}

	daily := make([]dailyActivityPoint, 0, days)
	total := 0
	for i := 0; i < days; i++ {
		dateKey := since.AddDate(0, 0, i).Format("2006-01-02")
		count := dailyCounts[dateKey]
		total += count
		daily = append(daily, dailyActivityPoint{Date: dateKey, Count: count})
	}

	repos := make([]repositoryActivity, 0, len(repoCounts))
	for name, commits := range repoCounts {
		repos = append(repos, repositoryActivity{Name: name, Commits: commits})
	}
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].Commits == repos[j].Commits {
			return repos[i].Name < repos[j].Name
		}
		return repos[i].Commits > repos[j].Commits
	})
	if len(repos) > 3 {
		repos = repos[:3]
	}

	average := 0
	if days > 0 {
		average = (total + days/2) / days
	}

	return &activityResponse{
		Daily: daily,
		Summary: activitySummary{
			TotalContributions: total,
			AverageDaily:       average,
		},
		RecentRepositories: repos,
	}, nil
}

func (h *ActivityHandler) persistActivities(userID, username string, res *activityResponse) {
	if h == nil || h.db == nil {
		return
	}

	records := make([]db.ActivityRecord, 0, len(res.Daily))
	for _, daily := range res.Daily {
		date, err := time.Parse("2006-01-02", daily.Date)
		if err != nil {
			continue
		}
		records = append(records, db.ActivityRecord{
			UserID:       userID,
			ActivityType: "push",
			Repository:   "",
			ActivityDate: date,
			CommitCount:  daily.Count,
		})
	}

	for _, repo := range res.RecentRepositories {
		records = append(records, db.ActivityRecord{
			UserID:       userID,
			ActivityType: "push",
			Repository:   repo.Name,
			ActivityDate: time.Now(),
			CommitCount:  repo.Commits,
		})
	}

	if err := db.SaveActivities(h.db, userID, records); err != nil {
		_ = fmt.Errorf("failed to save activities: %w", err)
	}
}

func (h *ActivityHandler) getFallbackActivities(userID string, days int) *activityResponse {
	if h == nil || h.db == nil {
		return nil
	}

	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	records, err := db.GetActivitiesByDateRange(h.db, userID, startDate, endDate)
	if err != nil || len(records) == 0 {
		return nil
	}

	dailyCounts := db.AggregateActivitiesByDate(records)
	repoCounts := db.AggregateActivitiesByRepository(records)

	daily := make([]dailyActivityPoint, 0, days)
	totalContributions := 0
	for i := 0; i < days; i++ {
		dateKey := startDate.AddDate(0, 0, i).Format("2006-01-02")
		count := dailyCounts[dateKey]
		totalContributions += count
		daily = append(daily, dailyActivityPoint{Date: dateKey, Count: count})
	}

	repos := make([]repositoryActivity, 0, len(repoCounts))
	for name, commits := range repoCounts {
		repos = append(repos, repositoryActivity{Name: name, Commits: commits})
	}
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].Commits == repos[j].Commits {
			return repos[i].Name < repos[j].Name
		}
		return repos[i].Commits > repos[j].Commits
	})
	if len(repos) > 3 {
		repos = repos[:3]
	}

	average := 0
	if days > 0 {
		average = (totalContributions + days/2) / days
	}

	return &activityResponse{
		Daily: daily,
		Summary: activitySummary{
			TotalContributions: totalContributions,
			AverageDaily:       average,
		},
		RecentRepositories: repos,
	}
}
