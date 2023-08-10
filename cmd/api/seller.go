package main

import (
	"errors"
	"net/http"

	"misarfeh.com/internal/data"
	"misarfeh.com/internal/validator"
)

func (app *application) registerSellerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FistName    string `json:"first_name"`
		LastName    string `json:"last_name"`
		Phone       string `json:"phone"`
		Email       string `json:"email,omitempty"`
		Password    string `json:"password"`
		MeliCode    string `json:"meli_code,omitempty"`
		MeliCartUrl string `json:"meli_cart_url,omitempty"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		FirstName: input.FistName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		Email:     input.Email,
	}

	seller := &data.Seller{
		MeliCode:    input.MeliCode,
		MeliCartUrl: input.MeliCartUrl,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if data.ValidateSeller(v, seller); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicatePhone):
			v.AddError("email", "a user with this phone number already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	seller.ID = user.ID

	err = app.models.Sellers.Insert(seller)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	output := struct {
		data.User
		MeliCode    string `json:"meli_code,omitempty"`
		MeliCartUrl string `json:"meli_cart_url,omitempty"`
	}{
		User:        *user,
		MeliCode:    seller.MeliCode,
		MeliCartUrl: seller.MeliCartUrl,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": output}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
