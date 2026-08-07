package news

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteNewsProvider struct {
	db *sql.DB
}

func NewSQLiteNewsProvider(dbPath string) (*SQLiteNewsProvider, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS news (
		id INTEGER PRIMARY KEY,
		published_at DATETIME,
		source TEXT,
		title TEXT,
		description TEXT,
		url TEXT,
		content TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_published ON news(published_at);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to init news schema: %w", err)
	}

	return &SQLiteNewsProvider{db: db}, nil
}

func (p *SQLiteNewsProvider) GetNewsBetween(start, end time.Time) ([]NewsItem, error) {
	query := `
		SELECT id, published_at, source, title, description, url, content
		FROM news
		WHERE published_at >= ? AND published_at <= ?
		ORDER BY published_at ASC
	`
	rows, err := p.db.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []NewsItem
	for rows.Next() {
		var item NewsItem
		if err := rows.Scan(
			&item.ID,
			&item.PublishedAt,
			&item.Source,
			&item.Title,
			&item.Description,
			&item.URL,
			&item.Content,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
