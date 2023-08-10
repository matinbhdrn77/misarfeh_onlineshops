package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"misarfeh.com/internal/validator"
)

type Shop struct {
	ID            int64     `json:"id"`
	CreatedAt     time.Time `json:"-"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	Year          int32     `json:"year,omitempty"`
	FollowerCount *int32    `json:"follower_count,omitempty"`
	Instagram     string    `json:"instagram,omitempty"`
	Telegram      string    `json:"telegram,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	LogoUrl       string    `json:"logo_url,omitempty"`
	Verified      bool      `json:"verified"`
	Rating        *float32  `json:"rating,omitempty"`
	RatingCount   int64     `json:"rating_count,omitempty"`
	Countries     []string  `json:"countries,omitempty"`
	Categories    []string  `json:"categories,omitempty"`
	ImgUrls       []string  `json:"img_urls,omitempty"`
	DeliveryTime  int8      `json:"delivery_time"`
}

func (s Shop) MarshalJSON() ([]byte, error) {
	var deliveryTime string

	if s.DeliveryTime != 0 {
		deliveryTime = fmt.Sprintf("%d هفته", s.DeliveryTime)
	}

	type ShopAlies Shop

	aux := struct {
		ShopAlies
		DeliveryTime string `json:"delivery_time,omitempty"`
	}{
		ShopAlies:    ShopAlies(s),
		DeliveryTime: deliveryTime,
	}

	return json.Marshal(aux)
}

func ValidateShop(v *validator.Validator, shop *Shop) {
	v.Check(shop.Title != "", "title", "must be provided")
	v.Check(len(shop.Title) <= 100, "title", "must not be more than 100 bytes long")

	v.Check(shop.Year != 0, "year", "must be provided")
	v.Check(shop.Year >= 1900, "year", "must be greater than 1888")
	v.Check(shop.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(len(shop.Description) <= 1000, "description", "must not be more than 1000 bytes long")

	v.Check(len(shop.Instagram) <= 100, "instagram", "must not be more than 100 bytes long")
	v.Check(len(shop.Telegram) <= 100, "telegram", "must not be more than 100 bytes long")

	v.Check(shop.Phone != "", "phone", "must be provided")
	v.Check(len(shop.Phone) == 11, "phone", "must be 11 bytes long")
	if shop.Phone != "" {
		v.Check(shop.Phone[0:2] == "09", "phone", "must be start with 09")
	}

	v.Check(shop.Countries != nil, "countries", "must be provided")
	v.Check(len(shop.Countries) >= 1, "countries", "must contain at least 1 countrie")
	v.Check(len(shop.Countries) <= 5, "countries", "must not contain more than 5 countries")
	v.Check(validator.Unique(shop.Countries), "countries", "must not contain duplicate values")

	v.Check(shop.Categories != nil, "categories", "must be provided")
	v.Check(len(shop.Categories) >= 1, "categories", "must contain at least 1 categorie")
	v.Check(len(shop.Categories) <= 5, "categories", "must not contain more than 5 categories")
	v.Check(validator.Unique(shop.Categories), "categories", "must not contain duplicate values")

	v.Check(shop.DeliveryTime != 0, "delivery_time", "must be provided")
	v.Check(shop.DeliveryTime >= 1, "delivery_time", "must be greater than 1 weak")
	v.Check(shop.DeliveryTime <= 100, "delivery_time", "must be lesser than 100 weak")
}

type ShopModel struct {
	DB *sql.DB
}

func (m ShopModel) GetAll(title string, verified bool, countries []string, filters Filters) ([]*Shop, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), shops.id, created_at, title, year, logo_url, delivery_time, 
			string_agg(DISTINCT countries.name, ',') AS country_names,
			string_agg(DISTINCT categories.name, ',') AS category_names
		FROM shops 
		FULL OUTER JOIN shops_countries ON shops.id = shops_countries.shop_id 
		LEFT JOIN countries ON shops_countries.country_id = countries.id
		FULL OUTER JOIN shops_categories ON shops.id = shops_categories.shop_id 
		LEFT JOIN categories ON shops_categories.category_id = categories.id
		WHERE (title ILIKE $1 OR instagram ILIKE $1 OR $1 = '')
		AND (countries.name = ANY($2) OR $2 = '{}')
		AND (categories.name = ANY($2) OR $2 = '{}')
		AND ($3 = false OR verified = $3)
		GROUP BY shops.id
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	title = fmt.Sprintf("%%%s%%", title)

	args := []interface{}{title, pq.Array(countries), verified, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	shops := []*Shop{}

	for rows.Next() {
		var shop Shop
		var country_names string
		var category_names string

		err := rows.Scan(
			&totalRecords,
			&shop.ID,
			&shop.CreatedAt,
			&shop.Title,
			&shop.Year,
			&shop.LogoUrl,
			&shop.DeliveryTime,
			&country_names,
			&category_names,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		shop.Countries = strings.Split(country_names, ",")
		shop.Categories = strings.Split(category_names, ",")

		shops = append(shops, &shop)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return shops, metadata, nil
}

func (m ShopModel) Insert(shop *Shop) error {
	query := `
		INSERT INTO shops (title, year, description, telegram, instagram, phone, logo_url, delivery_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	args := []interface{}{shop.Title, shop.Year, shop.Description, shop.Telegram,
		shop.Instagram, shop.Phone, shop.LogoUrl, shop.DeliveryTime}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&shop.ID, &shop.CreatedAt)
}

func (m ShopModel) Get(id int64) (*Shop, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, description, follower_count,
			telegram, instagram, phone, logo_url, rating, rating_count,
			verified, delivery_time
		FROM shops 
		WHERE id = $1`

	var shop Shop

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&shop.ID,
		&shop.CreatedAt,
		&shop.Title,
		&shop.Year,
		&shop.Description,
		&shop.FollowerCount,
		&shop.Telegram,
		&shop.Instagram,
		&shop.Phone,
		&shop.LogoUrl,
		&shop.Rating,
		&shop.RatingCount,
		&shop.Verified,
		&shop.DeliveryTime,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &shop, nil
}

func (m ShopModel) Update(shop *Shop) error {
	query := `
		UPDATE shops
		SET title = $1, year = $2, description = $3, telegram = $4,
	    	instagram = $5, phone = $6, logo_url = $7, delivery_time = $8
		WHERE id = $9
		RETURNING id`

	args := []interface{}{
		shop.Title,
		shop.Year,
		shop.Description,
		shop.Telegram,
		shop.Instagram,
		shop.Phone,
		shop.LogoUrl,
		shop.DeliveryTime,
		shop.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&shop.ID)
}

func (m ShopModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM shops
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

type MockShopModel struct{}

func (m MockShopModel) Insert(shop *Shop) error {
	return nil
}

func (m MockShopModel) Get(id int64) (*Shop, error) {
	return nil, nil
}

func (m MockShopModel) Update(shop *Shop) error {
	return nil
}

func (m MockShopModel) Delete(id int64) error {
	return nil
}

func (m MockShopModel) GetAll(title string, verified bool, countries []string, filters Filters) ([]*Shop, Metadata, error) {
	return nil, Metadata{}, nil
}
