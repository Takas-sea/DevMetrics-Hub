package models

import "time"

type User struct {
	ID        string    `db:"id"`
	GitHubID  int       `db:"github_id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	AvatarURL string    `db:"avatar_url"`
	Bio       string    `db:"bio"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Activity struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	ActivityType string    `db:"activity_type"`
	Repository   string    `db:"repository"`
	Description  string    `db:"description"`
	ActivityDate string    `db:"activity_date"`
	Count        int       `db:"count"`
	CreatedAt    time.Time `db:"created_at"`
}

type Session struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
