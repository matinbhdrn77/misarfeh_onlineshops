package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"misarfeh.com/internal/validator"
)

type Category struct {
	ID     int64  `josn:"id"`
	Name   string `josn:"name"`
	ImgUrl string `json:"img_url"`
}

func ValidateCategory(v *validator.Validator, category *Category) {
	v.Check(category.Name != "", "name", "must be provided")
	v.Check(len(category.Name) <= 100, "name", "must not be more than 100 bytes long")

	v.Check(category.ImgUrl != "", "img_url", "must be provided")
}

type CategoryModel struct {
	DB *sql.DB
}

func (m CategoryModel) GetAll(category *Category) ([]*Category, error) {
	query := `
		SELECT id, name, img_url
		FROM categories
		WHERE (id = $1 OR $1 = 0)
		AND (name = $2 OR $2 = '')`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, category.ID, category.Name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	categories := []*Category{}

	for rows.Next() {
		var category Category

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.ImgUrl,
		)

		if err != nil {
			return nil, err
		}

		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (m CategoryModel) Insert(category *Category) error {
	query := `
		INSERT INTO categories (name, img_url)
		VALUES ($1, $2)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, category.Name, category.ImgUrl).Scan(&category.ID)
}

func (m CategoryModel) GetOrInsert(names ...string) ([]*Category, error) {
	finalCategories := []*Category{}

	for _, name := range names {
		category := &Category{Name: name}
		savedCategories, err := m.GetAll(category)
		if err != nil {
			return nil, err
		}

		if len(savedCategories) == 0 {
			err = m.Insert(category)
			if err != nil {
				return nil, err
			}

			finalCategories = append(finalCategories, category)
		} else if len(savedCategories) == 1 {
			finalCategories = append(finalCategories, savedCategories...)
		} else {
			panic("there is 2 categories in db with same name")
		}
	}

	return finalCategories, nil
}

func (m CategoryModel) GetAllByShopID(id int64) ([]*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT categories.name 
		FROM shops_categories
		JOIN categories ON shops_categories.category_id = categories.id
		WHERE shop_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	categories := []*Category{}

	for rows.Next() {
		var category Category

		err := rows.Scan(
			&category.Name,
		)

		if err != nil {
			return nil, err
		}

		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (m CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name
		FROM categories 
		WHERE id = $1`

	var category Category

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &category, nil
}
