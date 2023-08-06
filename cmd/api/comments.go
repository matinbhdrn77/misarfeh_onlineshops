package main

import (
	"fmt"
	"net/http"
	"time"

	"misarfeh.com/internal/data"
	"misarfeh.com/internal/validator"
)

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text     string `json:"text"`
		Phone    string `json:"phone"`
		Username string `json:"username"`
		Rate     int8   `json:"rate"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comment := &data.Comment{
		Text:     input.Text,
		Phone:    input.Phone,
		Username: input.Username,
		Rate:     input.Rate,
	}

	v := validator.New()

	if data.ValidateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	comment := data.Comment{
		ID:        id,
		CreatedAt: time.Now(),
		Text:      "Good stuf",
		Phone:     "09379137805",
		Username:  "matin_nhd",
		Rate:      4,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCommentHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "list all comments")
}

func (app *application) EditeCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Edite comment %d\n", id)
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "delete comment %d\n", id)
}
