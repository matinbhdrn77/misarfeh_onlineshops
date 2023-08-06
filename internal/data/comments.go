package data

import (
	"time"

	"misarfeh.com/internal/validator"
)

type Comment struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Text      string    `json:"text"`
	Phone     string    `json:"phone"`
	Username  string    `json:"username"`
	Rate      int8      `json:"rate"`
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Text != "", "text", "must be provided")
	v.Check(len(comment.Text) <= 1000, "text", "must not be more than 1000 bytes long")

	v.Check(comment.Phone != "", "phone", "must be provided")
	v.Check(len(comment.Phone) == 11, "phone", "must be 11 bytes long")
	v.Check(comment.Phone[0:2] == "09", "phone", "must be start with 09")

	v.Check(comment.Username != "", "username", "must be provided")
	v.Check(len(comment.Username) <= 100, "username", "must not be more than 100 bytes long")

	v.Check(comment.Rate != 0, "rate", "must be provided")
	v.Check(comment.Rate >= 1, "rate", "must be greater than 0")
	v.Check(comment.Rate <= 5, "rate", "must be lesser than 5")
}
