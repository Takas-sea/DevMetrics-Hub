package db

import (
    "database/sql"
    "fmt"
    "os"
)

func Migrate(conn *sql.DB) error {
    schemaPath := "db/schema.sql"
    schema, err := os.ReadFile(schemaPath)
    if err != nil {
        return fmt.Errorf("failed to read schema file: %w", err)
    }

    _, err = conn.Exec(string(schema))
    if err != nil {
        return fmt.Errorf("failed to execute schema: %w", err)
    }

    return nil
}