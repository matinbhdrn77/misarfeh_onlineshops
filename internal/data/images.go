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

func (m ImageModel) GetAll(shop_id, product_id int64) ([]*Image, error) {
	if shop_id < 0 || product_id < 0 || (shop_id < 1 && product_id < 1) {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, url
		FROM images
		WHERE (shop_id = $1 OR $1 = 0)
		AND (product_id = $2 OR $2 = 0)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, shop_id, product_id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	images := []*Image{}

	for rows.Next() {
		var image Image

		err := rows.Scan(
			&image.ID,
			&image.Url,
		)

		if err != nil {
			return nil, err
		}

		images = append(images, &image)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}
