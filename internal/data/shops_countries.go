package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicateShopCountry = errors.New("duplicate shops_countries")
)

type ShopCountry struct {
	Shop_id    int64
	Country_id int64
}

type ShopCountryModel struct {
	DB *sql.DB
}

func (m ShopCountryModel) Insert(shopCountry *ShopCountry) error {
	query := `
		INSERT INTO shops_countries (shop_id, country_id)
		VALUES ($1, $2)
		RETURNING shop_id, country_id`

	args := []interface{}{shopCountry.Shop_id, shopCountry.Country_id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&shopCountry.Shop_id, &shopCountry.Country_id)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "shops_countries_pk"`:
			return ErrDuplicateShopCountry
		default:
			return err
		}
	}

	return nil
}

func (m ShopCountryModel) DeleteByShopID(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM shops_countries
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
