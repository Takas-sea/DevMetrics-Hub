package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Takas-sea/DevMetrics-Hub/utils"
)

type ActivityHandler struct{}

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

	res, err := h.fetchRecentActivities(claims.Username, days)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
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
