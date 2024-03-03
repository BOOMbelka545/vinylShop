package users

import (
	"context"
	"net/http"
	"vinylShop/config"
	"vinylShop/pkg/client/postgresql"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	cfg config.Properties
)

// Compare given password and stored
func isCredValid(givenPwd, storedPwd string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(storedPwd), []byte(givenPwd)); err != nil {
		log.Errorf("Cannot compare password: %v\n", err)
		return false
	}
	return true
}

// Authentification user
func AuthenticateUser(ctx context.Context, user postgresql.User) (postgresql.User, error) {
	var storedUser postgresql.User

	// Get password from BD
	sqlStatement := "SELECT password, isadmin FROM users WHERE email=($1)"
	err := db.QueryRow(context.Background(), sqlStatement, user.Email).Scan(&storedUser.Password, &storedUser.IsAdmin)
	if err == pgx.ErrNoRows {
		log.Errorf("Cannot get data from database: %v\n", err)
		return user, echo.NewHTTPError(http.StatusConflict, "User doesn't exist!")
	}

	if err != nil {
		log.Errorf("Unable to get data from db: %v", err)
		return storedUser, err
	}

	// if !isUserExist(context.Background(), user, db) {
	// 	return storedUser, echo.NewHTTPError(http.StatusNotFound, "User does not exist")
	// }

	if !isCredValid(user.Password, storedUser.Password) {
		return storedUser, echo.NewHTTPError(http.StatusUnauthorized, "Credentials invalid")
	}

	return postgresql.User{Email: user.Email, IsAdmin: storedUser.IsAdmin}, nil
}
