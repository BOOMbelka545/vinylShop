package cart

import (
	"context"
	"vinylShop/internal/db/postgresql"

	"github.com/labstack/gommon/log"
)

func AddProductToCart(cart postgresql.Cart) (int, error) {
	var id int
	db := postgresql.GetDB()
	sqlStatement := `
		INSERT INTO cart (user_id, product_id, count) VALUES ($1, $2, $3) RETURNING id
	`

	err := db.QueryRow(context.Background(), sqlStatement, cart.User_id, cart.Product_id, cart.Count).Scan(&id)
	if err != nil {
		log.Infof("Cannot insert into database: %v\n", err)
		return 0, err
	}

	return id, nil
}