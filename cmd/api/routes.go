package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/upload", app.uploadImagesHandler)

	fileServer := http.FileServer(http.Dir("./uploads/"))
	router.Handler(http.MethodGet, "/v1/images/*filepath", http.StripPrefix("/v1/images", fileServer))

	router.HandlerFunc(http.MethodGet, "/v1/shops", app.listShopsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/shops", app.createShopHandler)
	router.HandlerFunc(http.MethodGet, "/v1/shops/:id", app.showShopHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/shops/:id", app.updateShopHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/shops/:id", app.deleteShopHandler)

	router.HandlerFunc(http.MethodGet, "/v1/product/comments", app.listCommentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/product/comments", app.createCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product/comments/:id", app.showCommentHandler)
	router.HandlerFunc(http.MethodPut, "/v1/product/comments/:id", app.EditeCommentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/product/comments/:id", app.deleteCommentHandler)

	router.HandlerFunc(http.MethodGet, "/v1/product/categories", app.listCategoryHandler)
	router.HandlerFunc(http.MethodPost, "/v1/product/categories", app.createCategoryHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product/categories/:id", app.showCategoryHandler)
	router.HandlerFunc(http.MethodPut, "/v1/product/categories/:id", app.updateCategoryHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/product/categories/:id", app.deleteCategoryHandler)

	router.HandlerFunc(http.MethodGet, "/v1/products", app.listProductsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/products", app.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:id", app.showProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/products/:id", app.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/products/:id", app.deleteProductHandler)

	return app.recoverPanic(router)
}
