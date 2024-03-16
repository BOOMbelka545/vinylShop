package products

import (
	"context"
	"net/http"
	"vinylShop/internal/db/postgresql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func CreateProduct(product postgresql.Product) (int, error) {
	var id int
	db := postgresql.GetDB()
	sqlStatement := `
		INSERT INTO products
		(name, cost, artistname)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := db.QueryRow(context.Background(), sqlStatement, product.Name, product.Cost, product.ArtistName).Scan(&id)
	if err != nil {
		log.Errorf("Cannot insert product into database: %v\n", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return id, nil
}

func DeleteProduct(id int) (int, error) {
	db := postgresql.GetDB()
	sqlStatement := `
		DELETE FROM products
		WHERE id = ($1)
	`
	_, err := db.Exec(context.Background(), sqlStatement, id)
	if err != nil {
		log.Errorf("Cannot delete product from database: %v\n", err)
		return 0, err
	}

	return id, nil
}

func GetProduct(id int) (postgresql.Product, error) {
	var product postgresql.Product
	db := postgresql.GetDB()
	sqlStatement := `
		SELECT name, cost, artistname FROM products
		WHERE id = ($1)
	`

	err := db.QueryRow(context.Background(), sqlStatement, id).Scan(&product.Name, &product.Id, &product.ArtistName)
	if err != nil {
		log.Errorf("Cannot get product from database: %v\n", err)
		return product, err
	}

	return product, nil
}

func UpdateProduct(product postgresql.Product) (postgresql.Product, error) {
	db := postgresql.GetDB()
	var newProduct postgresql.Product
	sqlStatement := `
		UPDATE products
		SET name = $1,
			cost = $2,
			artistname = $3

		WHERE id = $4
		RETURNING name, cost, artistname
	`

	err := db.QueryRow(context.Background(), sqlStatement, product.Name, product.Cost, product.ArtistName, product.Id).Scan(&newProduct.Name, &newProduct.Name, &newProduct.ArtistName)
	if err != nil {
		log.Errorf("Cannot update product: %v\n", err)
		return newProduct, err
	}

	return newProduct, nil
}
