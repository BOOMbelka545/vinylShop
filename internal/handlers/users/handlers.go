package users

import (
	"context"
	"net/http"
	"time"
	"vinylShop/pkg/client/postgresql"

	"github.com/golang-jwt/jwt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	db = postgresql.GetDB()
)

func CreateToken(user postgresql.User, lifeTime time.Duration) (string, error) {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v", err)
	}
	claims := jwt.MapClaims{}

	claims["authorized"] = user.IsAdmin
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(lifeTime).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(cfg.JwtTokenSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetClaims(req *http.Request) (jwt.MapClaims, error) {
	jwtToken := req.Header["X-Access-Token"][0]
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtTokenSecret), nil
	})
	// Checking token validity
	if !token.Valid {
		log.Errorf("invalid token")
	}
	if err != nil {
		log.Infof("Cannot parse JWT token")
		return claims, echo.NewHTTPError(http.StatusInternalServerError, "Unable to parse JWT with claims")
	}

	return claims, nil
}

func UpdateUser(email interface{}, user postgresql.User) error {
	sqlStatement := `
		UPDATE users
		SET password = $2
		WHERE email = $1
		RETURNING id
	`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		log.Errorf("Unable to hash the password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There are some problems with server")
	}

	user.Password = string(hashedPassword)

	err = db.QueryRow(context.Background(), sqlStatement, email, user.Password).Scan(&user.Id)
	if err != nil {
		log.Errorf("Unable to update the password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There are some problems with server")
	}

	return nil
}
