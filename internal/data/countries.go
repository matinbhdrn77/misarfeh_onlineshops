package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"misarfeh.com/internal/validator"
)

type Country struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func ValidateCountry(v *validator.Validator, country *Country) {
	v.Check(country.Name != "", "name", "must be provided")
	v.Check(len(country.Name) <= 100, "name", "must not be more than 100 bytes long")
}

type CountryModel struct {
	DB *sql.DB
}

func (m CountryModel) GetAll(country *Country) ([]*Country, error) {
	query := `
		SELECT id, name
		FROM countries
		WHERE (id = $1 OR $1 = 0)
		AND (name = $2 OR $2 = '')`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, country.ID, country.Name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	countries := []*Country{}

	for rows.Next() {
		var country Country

		err := rows.Scan(
			&country.ID,
			&country.Name,
		)

		if err != nil {
			return nil, err
		}

		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

func (m CountryModel) Insert(country *Country) error {
	query := `
		INSERT INTO countries (name)
		VALUES ($1)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, country.Name).Scan(&country.ID)
}

func (m CountryModel) GetAllByShopID(id int64) ([]*Country, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT countries.name 
		FROM shops_countries 
		JOIN countries ON shops_countries.country_id = countries.id
		WHERE shop_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	countries := []*Country{}

	for rows.Next() {
		var country Country

		err := rows.Scan(
			&country.Name,
		)

		if err != nil {
			return nil, err
		}

		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

func (m CountryModel) GetOrInsert(names ...string) ([]*Country, error) {
	finalCountries := []*Country{}
	for _, name := range names {
		country := &Country{Name: name}
		savedCountries, err := m.GetAll(country)
		if err != nil {
			return nil, err
		}

		if len(savedCountries) == 0 {
			err = m.Insert(country)
			if err != nil {
				return nil, err
			}

			finalCountries = append(finalCountries, country)
		} else if len(savedCountries) == 1 {
			finalCountries = append(finalCountries, savedCountries...)
		} else {
			panic("there is 2 country in db with same name or id")
		}
	}

	return finalCountries, nil
}

func (m CountryModel) Get(id int64) (*Country, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name
		FROM countries 
		WHERE id = $1`

	var country Country

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&country.ID,
		&country.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &country, nil
}

func (m CountryModel) Update(country *Country) error {
	query := `
		UPDATE shops
		SET title = $1, year = $2, description = $3, telegram = $4,
	    	instagram = $5, phone = $6, logo_url = $7, delivery_time = $8
		WHERE id = $9
		RETURNING id`

	args := []interface{}{
		// shop.Title,
		// shop.Year,
		// shop.Description,
		// shop.Telegram,
		// shop.Instagram,
		// shop.Phone,
		// shop.LogoUrl,
		// shop.DeliveryTime,
		// shop.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(nil)
}

func (m CountryModel) Delete(id int64) error {
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

type MockCountryModel struct{}

func (m MockCountryModel) Insert(country *Country) error {
	return nil
}

func (m MockCountryModel) Get(id int64) (*Country, error) {
	return nil, nil
}

func (m MockCountryModel) Update(country *Country) error {
	return nil
}

func (m MockCountryModel) Delete(id int64) error {
	return nil
}

func (m MockCountryModel) GetAll(title string) ([]*Country, error) {
	return nil, nil
}

func GetAllByShopID(id int64) ([]*Country, error) {
	return nil, nil
}

func GetOrInsert(names []string) ([]*Country, error) {
	return nil, nil
}
