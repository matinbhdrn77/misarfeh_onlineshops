package main

import (
	"errors"
	"fmt"
	"net/http"

	"misarfeh.com/internal/data"
	"misarfeh.com/internal/validator"
)

func (app *application) createShopHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Year         int32    `json:"year"`
		Instagram    string   `json:"instagram"`
		Telegram     string   `json:"telegram"`
		Phone        string   `json:"phone"`
		Countries    []string `json:"countries"`
		Categories   []string `json:"categories"`
		DeliveryTime int8     `json:"delivery_time"`
		ImgUrls      []string `json:"img_urls"`
		LogoUrl      string   `json:"logo_url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create or return countries
	countries, err := app.models.Countries.GetOrInsert(input.Countries...)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create or return countries
	categories, err := app.models.Categories.GetOrInsert(input.Categories...)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	shop := &data.Shop{
		Title:        input.Title,
		Description:  input.Description,
		Year:         input.Year,
		Instagram:    input.Instagram,
		Telegram:     input.Telegram,
		Phone:        input.Phone,
		Countries:    input.Countries,
		Categories:   input.Categories,
		DeliveryTime: input.DeliveryTime,
		ImgUrls:      input.ImgUrls,
		LogoUrl:      input.LogoUrl,
	}

	v := validator.New()

	if data.ValidateShop(v, shop); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Shops.Insert(shop)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Insert ShopCountry
	for _, country := range countries {
		shopCountry := &data.ShopCountry{
			Shop_id:    shop.ID,
			Country_id: country.ID,
		}
		err = app.models.ShopCountry.Insert(shopCountry)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Insert ShopCategory
	for _, category := range categories {
		shopCategory := &data.ShopCategory{
			Shop_id:     shop.ID,
			Category_id: category.ID,
		}
		err = app.models.ShopCategory.Insert(shopCategory)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Insert images
	for _, url := range shop.ImgUrls {
		image := &data.Image{
			Url:    url,
			ShopID: &shop.ID,
		}

		err = app.models.Images.Insert(image)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/shops/%d", shop.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"shop": shop}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showShopHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	shop, err := app.models.Shops.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// retrieve list of countries
	countries, err := app.models.Countries.GetAllByShopID(shop.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return

	}

	for _, country := range countries {
		shop.Countries = append(shop.Countries, country.Name)
	}

	// retrieve list of categories
	categories, err := app.models.Categories.GetAllByShopID(shop.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return

	}

	for _, category := range categories {
		shop.Categories = append(shop.Categories, category.Name)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"shop": shop}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listShopsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title      string
		Instagram  string
		Countries  []string
		Categories []string
		Verified   bool
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Instagram = app.readString(qs, "instagram", "")
	input.Countries = app.readCSV(qs, "countries", []string{})
	input.Categories = app.readCSV(qs, "categories", []string{})
	input.Verified = app.readBool(qs, "verified", false)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "delivery_time", "-id", "-title", "-delivery_time"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	shops, metadata, err := app.models.Shops.GetAll(input.Title, input.Verified, input.Countries, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"shops": shops, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateShopHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	shop, err := app.models.Shops.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// fetch countries for shop
	countries, err := app.models.Countries.GetAllByShopID(shop.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	for _, country := range countries {
		shop.Countries = append(shop.Countries, country.Name)
	}

	// fetch categories for shop
	categories, err := app.models.Categories.GetAllByShopID(shop.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	for _, category := range categories {
		shop.Categories = append(shop.Categories, category.Name)
	}

	var input struct {
		Title        *string  `json:"title"`
		Description  *string  `json:"description"`
		Year         *int32   `json:"year"`
		Instagram    *string  `json:"instagram"`
		Telegram     *string  `json:"telegram"`
		Phone        *string  `json:"phone"`
		Countries    []string `json:"countries"`
		Categories   []string `json:"categories"`
		DeliveryTime *int8    `json:"delivery_time"`
		LogoUrl      *string  `json:"logo_url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	if input.Title != nil {
		shop.Title = *input.Title
	}

	if input.Description != nil {
		shop.Description = *input.Description
	}

	if input.Year != nil {
		shop.Year = *input.Year
	}

	if input.Instagram != nil {
		shop.Instagram = *input.Instagram
	}

	if input.Telegram != nil {
		shop.Telegram = *input.Telegram
	}

	if input.Phone != nil {
		shop.Phone = *input.Phone
	}

	if input.Countries != nil {
		shop.Countries = input.Countries
	}

	if input.Categories != nil {
		shop.Categories = input.Categories
	}

	v := validator.New()

	if data.ValidateShop(v, shop); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Shops.Update(shop)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Delete all shops_countries row for this shop
	if len(shop.Countries) != 0 {
		err = app.models.ShopCountry.DeleteByShopID(shop.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
				return
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	// Create or return country_id
	countries, err = app.models.Countries.GetOrInsert(input.Countries...)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Insert ShopCountry
	for _, country := range countries {
		shopCountry := &data.ShopCountry{
			Shop_id:    shop.ID,
			Country_id: country.ID,
		}
		err = app.models.ShopCountry.Insert(shopCountry)
		if err != nil {

			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Delete all shops_categories row for this shop
	if len(shop.Categories) != 0 {
		err = app.models.ShopCategory.DeleteByShopID(shop.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
				return
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	// Create or return category_id
	categories, err = app.models.Categories.GetOrInsert(input.Categories...)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Insert ShopCategory
	for _, category := range categories {
		shopCategory := &data.ShopCategory{
			Shop_id:     shop.ID,
			Category_id: category.ID,
		}
		err = app.models.ShopCategory.Insert(shopCategory)
		if err != nil {

			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"shop": shop}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteShopHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Shops.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "shop succesfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
