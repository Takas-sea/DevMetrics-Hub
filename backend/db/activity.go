package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ActivityRecord struct {
	UserID          string
	ActivityType    string
	Repository      string
	ActivityDate    time.Time
	CommitCount     int
}

func SaveActivities(conn *sql.DB, userID string, activities []ActivityRecord) error {
	if conn == nil || len(activities) == 0 {
		return nil
	}

	for _, act := range activities {
		id := uuid.New().String()
		query := `
			INSERT INTO activities (id, user_id, activity_type, repository, activity_date, count, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (user_id, activity_type, repository, activity_date)
			DO UPDATE SET count = EXCLUDED.count
		`
		_, err := conn.Exec(query,
			id,
			userID,
			act.ActivityType,
			act.Repository,
			act.ActivityDate,
			act.CommitCount,
			time.Now())
		if err != nil {
			return fmt.Errorf("failed to upsert activity: %w", err)
		}
	}
	return nil
}

func GetActivitiesByDateRange(conn *sql.DB, userID string, startDate time.Time, endDate time.Time) ([]ActivityRecord, error) {
	if conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT user_id, activity_type, repository, activity_date, count
		FROM activities
		WHERE user_id = $1 AND activity_date >= $2 AND activity_date <= $3
		ORDER BY activity_date DESC, repository ASC
	`

	rows, err := conn.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var records []ActivityRecord
	for rows.Next() {
		var r ActivityRecord
		if err := rows.Scan(&r.UserID, &r.ActivityType, &r.Repository, &r.ActivityDate, &r.CommitCount); err != nil {
			return nil, fmt.Errorf("failed to scan activity row: %w", err)
		}
		records = append(records, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity rows: %w", err)
	}

	return records, nil
}

func AggregateActivitiesByDate(records []ActivityRecord) map[string]int {
	aggregated := make(map[string]int)
	for _, r := range records {
		dateKey := r.ActivityDate.Format("2006-01-02")
		aggregated[dateKey] += r.CommitCount
	}
	return aggregated
}

func AggregateActivitiesByRepository(records []ActivityRecord) map[string]int {
	aggregated := make(map[string]int)
	for _, r := range records {
		if r.Repository != "" {
			aggregated[r.Repository] += r.CommitCount
		}
	}
	return aggregated
}
