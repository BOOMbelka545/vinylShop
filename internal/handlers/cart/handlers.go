package cart

import (
	"context"
	"net/http"
	"strconv"
	"vinylShop/internal/db/postgresql"
	u "vinylShop/internal/handlers/users"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func AddProductToCart(cart postgresql.Cart) error {
	db := postgresql.GetDB()
	sqlStatement := `
		INSERT INTO cart (user_id, product_id, count) VALUES ($1, $2, $3)
	`

	_, err := db.Exec(context.Background(), sqlStatement, cart.User_id, cart.Product_id, cart.Count)
	if err != nil {
		log.Infof("Cannot insert into database: %v\n", err)
		return err
	}

	return nil
}

func SubExistsProduct(c echo.Context) error {
	var count int
	var userId int
	sqlGetUserId := `
		SELECT (id) FROM users
		WHERE email = ($1)
	`
	sqlUpdateValue := `
		UPDATE cart SET count = count - 1
		WHERE user_id = ($1) AND product_id = ($2)
		RETURNING count
	`
	sqlDeleteProduct := `
		DELETE FROM cart 
		WHERE user_id = ($1) AND product_id = ($2)
	`

	db := postgresql.GetDB()

	productId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Cannot convert id from string to int: %v", err)
		return err
	}

	claims, err := u.GetClaims(c.Request())
	if err != nil {
		log.Errorf("Cannot get claims from request: %v", err)
		return err
	}
	email := claims["email"]

	err = db.QueryRow(context.Background(), sqlGetUserId, email).Scan(&userId)
	if err != nil {
		log.Errorf("Cannot get user id from bd: %v", err)
		return err
	}

	err = db.QueryRow(context.Background(), sqlUpdateValue, userId, productId).Scan(&count)
	if err != nil {
		log.Errorf("Cannot update value: %v", err)
		return err
	}

	if count == 0 {
		_, err = db.Exec(context.Background(), sqlDeleteProduct, userId, productId)
		if err != nil {
			log.Errorf("Cannot Delete product: %v", err)
			return err
		}
	}

	return nil
}

func GetCart(c echo.Context) (postgresql.Cart, error) {
	var cart postgresql.Cart
	sqlStatement := `
		SELECT (id) FROM users
		WHERE email = ($1)
	`
	db := postgresql.GetDB()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Cannot convert id from string to int: %v", err)
		return cart, err
	}
	cart.Product_id = id
	if err := c.Bind(&cart.Count); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return cart, echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	claims, err := u.GetClaims(c.Request())
	if err != nil {
		log.Errorf("Cannot get claims from request: %v", err)
		return cart, err
	}
	email := claims["email"]

	err = db.QueryRow(context.Background(), sqlStatement, email).Scan(&cart.User_id)
	if err != nil {
		log.Errorf("Cannot get user id from bd: %v", err)
		return cart, err
	}

	return cart, nil
}

func AddExistsProduct(c echo.Context) error {
	var userId int
	sqlGetUserId := `
		SELECT (id) FROM users
		WHERE email = ($1)
	`

	sqlUpdateValue := `
		UPDATE cart SET count = count + 1
		WHERE user_id = ($1) AND product_id = ($2)
	`

	db := postgresql.GetDB()

	productId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	claims, err := u.GetClaims(c.Request())
	if err != nil {
		return err
	}
	email := claims["email"]

	err = db.QueryRow(context.Background(), sqlGetUserId, email).Scan(&userId)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), sqlUpdateValue, userId, productId)
	if err != nil {
		log.Errorf("Cannot update value: %v", err)
		return err
	}

	return nil
}

func Payment(c echo.Context) error {
	var userId int
	sqlStatement := `
		DELETE FROM cart 
		WHERE user_id = $1
	`
	sqlGetUserId := `
		SELECT (id) FROM users
		WHERE email = ($1)
	`	
	db := postgresql.GetDB()

	claims, err := u.GetClaims(c.Request())
	if err != nil {
		log.Errorf("Cannot get claims from request: %v", err)
		return err
	}
	email := claims["email"]

	err = db.QueryRow(context.Background(), sqlGetUserId, email).Scan(&userId)
	if err != nil {
		log.Errorf("Cannot get user id from bd: %v", err)
		return err
	}

	_, err = db.Exec(context.Background(), sqlStatement, userId)
	if err != nil {
		log.Errorf("Cannot delete users cart from bd: %v", err)
		return err
	}

	return nil
}