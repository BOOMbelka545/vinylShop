package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
	"vinylShop/config"
	p "vinylShop/internal/handlers/products"
	u "vinylShop/internal/handlers/users"
	"vinylShop/pkg/client/postgresql"

	"github.com/golang-jwt/jwt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gopkg.in/go-playground/validator.v9"
)

type CustomValidator struct {
	validator *validator.Validate
}

var (
	v   = validator.New()
	cfg config.Properties
)

const (
	accessTokenMaxAge  = 10 * time.Minute
	refreshTokenMaxAge = 5 * time.Hour
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// Create user handler
func CreateUser(c echo.Context) error {
	// user := make(map[string]interface{}) OR  var user User OR user = User{}
	// json.NewDecoder(c.Request().Body).Decode(&user)  -- another way to get request body

	// Get user from request body
	user := postgresql.User{}
	if err := c.Bind(&user); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	// Validate data
	c.Echo().Validator = &CustomValidator{validator: v}
	if err := c.Validate(user); err != nil {
		return err
	}

	// Insert the new user in DB
	newUserId, err := u.InsertUser(context.Background(), user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newUserId)
}

// Authenticate user handler
func AuthnUser(c echo.Context) error {
	// Get user from request body
	user := postgresql.User{}
	if err := c.Bind(&user); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	// Authenticate user
	user, err := u.AuthenticateUser(context.Background(), user)
	if err != nil {
		log.Errorf("Unable to authenticate to db.")
		return err
	}

	// Create access token
	accessToken, er := u.CreateToken(user, accessTokenMaxAge)
	if er != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}
	// Set access token to response header
	c.Response().Header().Set("X-Access-Token", accessToken)

	// Create refresh token
	refreshToken, er := u.CreateToken(user, refreshTokenMaxAge)
	if er != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}
	// Set refresh token to response header
	c.Response().Header().Set("X-Refresh-Token", refreshToken)

	return c.JSON(http.StatusOK, user)
}

// Update password
func UpdateUser(c echo.Context) error {
	claims, err := u.GetClaims(c.Request())
	if err != nil {
		return err
	}

	email := claims["email"]

	// Bind request to User struct
	user := postgresql.User{}
	if err = c.Bind(&user); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	// Update user
	err = u.UpdateUser(email, user)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "The user has been updated")
}

// Refresh Token
func RefreshToken(c echo.Context) error {
	var user postgresql.User

	db := postgresql.GetDB()

	// Read config from environment
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v\n", err)
	}

	// Get refresh token from request body
	refreshToken := c.Request().Header["X-Refresh-Token"][0]
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtTokenSecret), nil
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to parse JWT with claims")
	}

	// Checking token validity
	if !token.Valid {
		log.Errorf("invalid token")
	}

	email := claims["email"]

	// SQL statement to get user from db
	sqlStatement := `
		SELECT (id_u, email, password, isadmin) FROM users
		WHERE email = ($1);
	`

	// SQL request
	err = db.QueryRow(context.Background(), sqlStatement, email).Scan(&user)
	if err != nil {
		log.Errorf("Unable to get the user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There are some problems with server")
	}

	// Create new access token
	newAccessToken, err := u.CreateToken(user, accessTokenMaxAge)
	if err != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}

	// Set new access token to response request
	c.Response().Header().Set("X-Access-Token", newAccessToken)

	// Create new refresh token
	refreshToken, err = u.CreateToken(user, refreshTokenMaxAge)
	if err != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}
	// Set new refresh token to response header
	c.Response().Header().Set("X-Refresh-Token", refreshToken)

	return nil
}

// Create product
func CreateProduct(c echo.Context) error {
	var product postgresql.Product
	if err := c.Bind(&product); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	id, err := p.CreateProduct(product)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, id)
}

// Delete product
func DeleteProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	id, err = p.DeleteProduct(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, id)
}

// Get product
func GetProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	product, err := p.GetProduct(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, "Not Found")
		}
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, product)
}

// Update product
func UpdateProduct(c echo.Context) error {
	var product postgresql.Product
	if err := c.Bind(&product); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	product, err := p.UpdateProduct(product)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, product)
}