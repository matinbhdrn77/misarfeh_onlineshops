package data

import (
	"time"
)

type Image struct {
	ID        int64     `json:"id"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
