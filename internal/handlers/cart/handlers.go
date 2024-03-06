package cart

import (
	u "vinylShop/internal/handlers/users"
	"vinylShop/pkg/client/postgresql"
)

func AddProductToCart(id int) (int, error) {
	db := postgresql.GetDB()
	sqlStatement := `
		INSERT INTO cart (user_id, product_id, count)
		VALUES (&1, &2, &3)
	`

	
}