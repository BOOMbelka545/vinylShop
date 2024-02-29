package products

import (
	"context"
	"net/http"
	"vinylShop/pkg/client/postgresql"

	"github.com/labstack/echo/v4"
)

func CreateProduct(product postgresql.Product) (int, error) {
	var id int
	db := postgresql.GetDB()
	sqlStatement := `
		INSERT INTO products
		(name, cost, artistname)
		VALUES ($1, $2, $3)
		RETURNING id_p
	`
	err := db.QueryRow(context.Background(), sqlStatement, product.Name, product.Cost, product.ArtistName).Scan(&id)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return id, nil
}

// func GetProduct() (product, error) {
// }
//

// func UpdateProduct() (product, error) {
// }

func DeleteProduct() error {
	return nil
}
