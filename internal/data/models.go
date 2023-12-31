package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Shops interface {
		Insert(shop *Shop) error
		Get(id int64) (*Shop, error)
		Update(shop *Shop) error
		Delete(id int64) error
		GetAll(title string, verified bool, countries []string, filters Filters) ([]*Shop, Metadata, error)
	}
	Countries interface {
		Insert(country *Country) error
		Get(id int64) (*Country, error)
		Update(country *Country) error
		Delete(id int64) error
		GetAll(country *Country) ([]*Country, error)
		GetAllByShopID(id int64) ([]*Country, error)
		GetOrInsert(countries ...string) ([]*Country, error)
	}
	ShopCountry interface {
		Insert(shopCountry *ShopCountry) error
		DeleteByShopID(id int64) error
	}
	ShopCategory interface {
		Insert(shopCategory *ShopCategory) error
		DeleteByShopID(id int64) error
	}
	Products interface {
		Insert(product *Product) error
		Get(id int64) (*Product, error)
		Update(product *Product) error
		Delete(id int64) error
		GetAll(title, brand string, shop_id, country_id int64, filters Filters) ([]*Shop, Metadata, error)
	}
	Categories interface {
		Insert(category *Category) error
		Get(id int64) (*Category, error)
		// Update(country *Country) error
		// Delete(id int64) error
		GetAll(category *Category) ([]*Category, error)
		GetAllByShopID(id int64) ([]*Category, error)
		GetOrInsert(categories ...string) ([]*Category, error)
	}
	Users interface {
		Insert(user *User) error
		GetByEmailPhone(emailPhone string) (*User, error)
		Update(user *User) error
	}
	Sellers interface {
		Insert(seller *Seller) error
		Get(id int64) (*Seller, error)
		Update(seller *Seller) error
	}
	Images interface {
		Insert(image *Image) error
		GetAll(shop_id, product_id int64) ([]*Image, error)
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Shops:        ShopModel{DB: db},
		Countries:    CountryModel{DB: db},
		ShopCountry:  ShopCountryModel{DB: db},
		ShopCategory: ShopCategoryModel{DB: db},
		Products:     ProductModel{DB: db},
		Categories:   CategoryModel{DB: db},
		Users:        UserModel{DB: db},
		Sellers:      SellerModel{DB: db},
		Images:       ImageModel{DB: db},
	}
}

func NewMockModels() Models {
	return Models{
		Shops:     MockShopModel{},
		Countries: CountryModel{},
	}
}
