package users

import (
	"context"
	"net/http"
	"vinylShop/pkg/client/postgresql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

// find user to check Is user exist
func IsUserExist(ctx context.Context, user postgresql.User) bool {
	var isUser bool
	sqlStatement := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email=($1))
	`
	err := db.QueryRow(ctx, sqlStatement, user.Email).Scan(&isUser)
	if err != nil {
		log.Fatalf("Unable to check your email: %v \n", err)
	}

	return isUser
}

// Register a user
func InsertUser(ctx context.Context, user postgresql.User) (int, error) {
	if IsUserExist(context.Background(), user) {
		return 0, echo.NewHTTPError(http.StatusConflict, "User already exist!")
	}

	sqlStatement := "INSERT INTO users (email, password, isAdmin) VALUES ($1, $2, $3) RETURNING id"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		log.Errorf("Unable to hash the password: %v", err)
		return user.Id, err
	}
	user.Password = string(hashedPassword)

	err = db.QueryRow(ctx, sqlStatement, user.Email, user.Password, user.IsAdmin).Scan(&user.Id)
	if err != nil {
		log.Errorf("Cannot insert user into database: %v\n", err)
		return user.Id, err
	}

	return user.Id, nil
}
