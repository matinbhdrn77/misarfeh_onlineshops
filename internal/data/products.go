package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"misarfeh.com/internal/validator"
)

type Product struct {
	ID          int64     `json:"id"`
	ShopID      int64     `json:"-"`
	CategoryID  int64     `json:"-"`
	Category    string    `json:"category"`
	CountryID   int64     `json:"-"`
	Country     string    `json:"country"`
	CreatedAt   time.Time `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float32   `json:"price"`
	SalePrice   int64     `json:"sale_price"`
	Off         int32     `json:"Off"`
	Brand       string    `json:"brand"`
	ImgUrls     []string  `json:"img_urls"`
}

func (p Product) MarshalJSON() ([]byte, error) {
	var off string

	if p.Off != 0 {
		off = fmt.Sprintf("%d%%", p.Off)
	}

	type ProductAlies Product

	aux := struct {
		ProductAlies
		Off string `json:"off,omitempty"`
	}{
		ProductAlies: ProductAlies(p),
		Off:          off,
	}

	return json.Marshal(aux)
}

func ValidateProduct(v *validator.Validator, product *Product) {
	v.Check(product.Name != "", "name", "must be provided")
	v.Check(len(product.Name) <= 100, "name", "must not be more than 100 bytes long")

	v.Check(product.Price >= float32(0), "price", "must be greater than zero")

	v.Check(product.SalePrice != 0, "sale_price", "must be provided")

	v.Check(product.Off < 100, "off", "must not be greater than 100")
	v.Check(product.Off >= 0, "off", "must not be negetive number")

	v.Check(len(product.Description) <= 1000, "description", "must not be more than 1000 bytes long")

	v.Check(product.Brand != "", "brand", "must be provided")
	v.Check(len(product.Brand) <= 100, "brand", "must not be more than 100 bytes long")

	v.Check(product.Category != "", "category", "must be provided")
	v.Check(len(product.Category) <= 100, "category", "must not be more than 100 bytes long")

	v.Check(product.Country != "", "country", "must be provided")
	v.Check(len(product.Country) <= 100, "country", "must not be more than 100 bytes long")

	v.Check(product.ImgUrls != nil, "img_urls", "must be provided")
	v.Check(len(product.ImgUrls) >= 1, "img_urls", "must contain at least 1 img")
	v.Check(len(product.ImgUrls) <= 5, "img_urls", "must not contain more than 5 img")
	v.Check(validator.Unique(product.ImgUrls), "img_urls", "must not contain duplicate values")
}

type ProductModel struct {
	DB *sql.DB
}

func (m ProductModel) GetAll(title, brand string, shop_id, country_id int64, filters Filters) ([]*Shop, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), shops.id, created_at, title, year, logo_url, delivery_time
		FROM shops 
		JOIN shops_countries ON shops.id = shops_countries.shop_id 
		JOIN countries ON shops_countries.country_id = countries.id
		WHERE (title ILIKE $1 OR instagram ILIKE $1 OR $1 = '')
		AND (name = ANY($2) OR $2 = '{}')
		GROUP BY shops.id
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	title = fmt.Sprintf("%%%s%%", title)

	args := []interface{}{title, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	shops := []*Shop{}

	for rows.Next() {
		var shop Shop

		err := rows.Scan(
			&totalRecords,
			&shop.ID,
			&shop.CreatedAt,
			&shop.Title,
			&shop.Year,
			&shop.LogoUrl,
			&shop.DeliveryTime,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		shops = append(shops, &shop)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return shops, metadata, nil
}

func (m ProductModel) Insert(product *Product) error {
	query := `
		INSERT INTO products (shop_id, category_id, country_id,
			name, description, price, sale_price, off, brand)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	args := []interface{}{4, product.CategoryID, product.CountryID,
		product.Name, product.Description, product.Price, product.SalePrice,
		product.Off, product.Brand}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID, &product.CreatedAt)
}

func (m ProductModel) Get(id int64) (*Product, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, shop_id, category_id, country_id, name,
			description, price, sale_price, off, brand
		FROM products 
		WHERE id = $1`

	var product Product

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.ShopID,
		&product.CategoryID,
		&product.CountryID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SalePrice,
		&product.Off,
		&product.Brand,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &product, nil
}

func (m ProductModel) Update(product *Product) error {
	query := `
		UPDATE products
		SET name = $1, category_id = $2, country_id = $3, description = $4,
	    	price = $5, sale_price = $6, off = $7, brand = $8
		WHERE id = $9
		RETURNING id`

	args := []interface{}{
		product.Name,
		product.CategoryID,
		product.CountryID,
		product.Description,
		product.Price,
		product.SalePrice,
		product.Off,
		product.Brand,
		product.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID)
}

func (m ProductModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM products
		WHERE id = $1`

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

type MockProductModel struct{}

func (m MockProductModel) Insert(product *Product) error {
	return nil
}

func (m MockProductModel) Get(id int64) (*Product, error) {
	return nil, nil
}

func (m MockProductModel) Update(product *Product) error {
	return nil
}

func (m MockProductModel) Delete(id int64) error {
	return nil
}

func (m MockProductModel) GetAll(name, brand string, shop_id, country_id int64, filters Filters) ([]*Shop, Metadata, error) {
	return nil, Metadata{}, nil
}
