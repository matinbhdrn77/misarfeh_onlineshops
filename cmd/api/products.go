package main

import (
	"errors"
	"fmt"
	"net/http"

	"misarfeh.com/internal/data"
	"misarfeh.com/internal/validator"
)

// Todo : add real shop_id to created products by JWT, and img

func (app *application) createProductHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Category    string   `json:"category"`
		Country     string   `json:"country"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       float32  `json:"price,omitempty"`
		SalePrice   int64    `json:"sale_price"`
		Off         int32    `json:"off,omitempty"`
		Brand       string   `json:"brand"`
		ImgUrls     []string `json:"img_urls"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		Name:        input.Name,
		Country:     input.Country,
		Category:    input.Category,
		Description: input.Description,
		Price:       input.Price,
		SalePrice:   input.SalePrice,
		Off:         input.Off,
		Brand:       input.Brand,
		ImgUrls:     input.ImgUrls,
	}

	v := validator.New()

	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create or return Country
	countries, err := app.models.Countries.GetOrInsert(input.Country)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create or return Category
	categories, err := app.models.Categories.GetOrInsert(input.Category)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Update product
	product.CountryID = countries[0].ID
	product.CategoryID = categories[0].ID

	err = app.models.Products.Insert(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/products/%d", product.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"product": product}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	product, err := app.models.Products.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// fetch country name and add to product
	country, err := app.models.Countries.Get(product.CountryID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			product.Country = "country not found"
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	} else {
		product.Country = country.Name
	}

	// fetch category name and add to product
	category, err := app.models.Categories.Get(product.CategoryID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			product.Category = "category not found"
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	} else {
		product.Category = category.Name
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listProductsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string
		Brand      string
		categories []string
		Countries  []string
		Sale_Price []int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "title", "")
	input.Brand = app.readString(qs, "instagram", "")
	input.Countries = app.readCSV(qs, "countries", []string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "delivery_time", "-id", "-title", "-delivery_time"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	shops, metadata, err := app.models.Shops.GetAll(input.Name, false, input.Countries, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"shops": shops, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	product, err := app.models.Products.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Category    *string  `json:"category"`
		Country     *string  `json:"country"`
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float32 `json:"price,omitempty"`
		SalePrice   *int64   `json:"sale_price"`
		Off         *int32   `json:"off,omitempty"`
		Brand       *string  `json:"brand"`
		ImgUrls     []string `json:"img_urls"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create or return Country
	countries, err := app.models.Countries.GetOrInsert(*input.Country)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create or return Category
	categories, err := app.models.Categories.GetOrInsert(*input.Category)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if input.Category != nil {
		product.CategoryID = categories[0].ID
		product.Category = categories[0].Name
	}

	if input.Country != nil {
		product.CountryID = countries[0].ID
		product.Country = countries[0].Name
	}

	if input.Name != nil {
		product.Name = *input.Name
	}

	if input.Description != nil {
		product.Description = *input.Description
	}

	if input.Price != nil {
		product.Price = *input.Price
	}

	if input.SalePrice != nil {
		product.SalePrice = *input.SalePrice
	}

	if input.Off != nil {
		product.Off = *input.Off
	}

	if input.Brand != nil {
		product.Brand = *input.Brand
	}

	if input.ImgUrls != nil {
		product.ImgUrls = input.ImgUrls
	}

	v := validator.New()

	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Products.Update(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Products.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "product succesfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
