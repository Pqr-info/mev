package news

import (
	"time"
)

type NewsItem struct {
	ID          int       `json:"id"`
	PublishedAt time.Time `json:"published_at"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Content     string    `json:"content"`
}

type NewsProvider interface {
	GetNewsBetween(start, end time.Time) ([]NewsItem, error)
}
