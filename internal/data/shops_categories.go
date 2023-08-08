package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicateShopCategory = errors.New("duplicate shops_categories")
)

type ShopCategory struct {
	Shop_id     int64
	Category_id int64
}

type ShopCategoryModel struct {
	DB *sql.DB
}

func (m ShopCategoryModel) Insert(shopCategory *ShopCategory) error {
	query := `
		INSERT INTO shops_categories (shop_id, category_id)
		VALUES ($1, $2)
		RETURNING shop_id, category_id`

	args := []interface{}{shopCategory.Shop_id, shopCategory.Category_id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&shopCategory.Shop_id, &shopCategory.Category_id)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "shops_categories_pk"`:
			return ErrDuplicateShopCountry
		default:
			return err
		}
	}

	return nil
}

func (m ShopCategoryModel) DeleteByShopID(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM shops_categories
		WHERE shop_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
