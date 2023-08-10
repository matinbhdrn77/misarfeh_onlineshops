package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"misarfeh.com/internal/validator"
)

type Seller struct {
	ID          int64  `json:"id"`
	MeliCode    string `json:"meli_code"`
	MeliCartUrl string `json:"meli_cart_url"`
}

type SellerModel struct {
	DB *sql.DB
}

func (m SellerModel) Insert(seller *Seller) error {
	query := `
		INSERT INTO sellers (id, meli_code, meli_cart_url)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []interface{}{seller.ID, seller.MeliCode, seller.MeliCartUrl}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&seller.ID)
}

func (m SellerModel) Get(id int64) (*Seller, error) {
	query := `
		SELECT id, meli_code, meli_cart_url
		FROM sellers
		WHERE id = $1`

	var seller Seller

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&seller.ID,
		&seller.MeliCode,
		&seller.MeliCartUrl,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &seller, nil
}

func (m SellerModel) Update(seller *Seller) error {
	query := `
	UPDATE sellers
	SET meli_code = $1, meli_cart_url = $2
	WHERE id = $3
	RETURNING id`

	args := []interface{}{
		seller.MeliCode, seller.MeliCartUrl, seller.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&seller.ID)
}

func ValidateSeller(v *validator.Validator, seller *Seller) {
	v.Check(len(seller.MeliCode) <= 20, "meli_code", "must not be more than 20 bytes long")
	v.Check(validator.Matches(seller.MeliCode, validator.MeliCodeRX) || seller.MeliCode == "", "meli_code", "must be a valid meli code")
}
