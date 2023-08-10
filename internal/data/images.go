package data

import (
	"context"
	"database/sql"
	"time"
)

type Image struct {
	ID        int64  `json:"id"`
	Url       string `json:"url"`
	ProductID *int64 `json:"product_id"`
	ShopID    *int64 `json:"shop_id"`
}

type ImageModel struct {
	DB *sql.DB
}

func (m ImageModel) Insert(image *Image) error {
	query := `
		INSERT INTO images (url, product_id, shop_id)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []interface{}{image.Url, image.ProductID, image.ShopID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&image.ID)
}
