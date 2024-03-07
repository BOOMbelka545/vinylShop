package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
	"vinylShop/config"
	crt "vinylShop/internal/handlers/cart"
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

func CreateUser(c echo.Context) error {
	// user := make(map[string]interface{}) OR  var user User OR user = User{}
	// json.NewDecoder(c.Request().Body).Decode(&user)  -- another way to get request body

	user := postgresql.User{}
	if err := c.Bind(&user); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	c.Echo().Validator = &CustomValidator{validator: v}
	if err := c.Validate(user); err != nil {
		return err
	}

	newUserId, err := u.InsertUser(context.Background(), user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newUserId)
}

func AuthnUser(c echo.Context) error {
	user := postgresql.User{}
	if err := c.Bind(&user); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	user, err := u.AuthenticateUser(context.Background(), user)
	if err != nil {
		log.Errorf("Unable to authenticate to db.")
		return err
	}

	accessToken, er := u.CreateToken(user, accessTokenMaxAge)
	if er != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}
	c.Response().Header().Set("X-Access-Token", accessToken)

	refreshToken, er := u.CreateToken(user, refreshTokenMaxAge)
	if er != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}
	c.Response().Header().Set("X-Refresh-Token", refreshToken)

	return c.JSON(http.StatusOK, user)
}

func UpdateUser(c echo.Context) error {
	claims, err := u.GetClaims(c.Request())
	if err != nil {
		return err
	}

	email := claims["email"]

	user := postgresql.User{}
	if err = c.Bind(&user); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	err = u.UpdateUser(email, user)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "The user has been updated")
}

func RefreshToken(c echo.Context) error {
	var user postgresql.User

	db := postgresql.GetDB()

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v\n", err)
	}

	refreshToken := c.Request().Header["X-Refresh-Token"][0]
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtTokenSecret), nil
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to parse JWT with claims")
	}

	if !token.Valid {
		log.Errorf("invalid token")
	}

	email := claims["email"]

	sqlStatement := `
		SELECT (id_u, email, password, isadmin) FROM users
		WHERE email = ($1);
	`

	err = db.QueryRow(context.Background(), sqlStatement, email).Scan(&user)
	if err != nil {
		log.Errorf("Unable to get the user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There are some problems with server")
	}

	newAccessToken, err := u.CreateToken(user, accessTokenMaxAge)
	if err != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}

	c.Response().Header().Set("X-Access-Token", newAccessToken)

	refreshToken, err = u.CreateToken(user, refreshTokenMaxAge)
	if err != nil {
		log.Errorf("Unable to generate the token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}
	c.Response().Header().Set("X-Refresh-Token", refreshToken)

	return nil
}

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

func UpdateProduct(c echo.Context) error {
	var product postgresql.Product
	if err := c.Bind(&product); err != nil {
		log.Errorf("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	product, err := p.UpdateProduct(product)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, product)
}

func AddProductToCart(c echo.Context) error {
	var cart postgresql.Cart
	sqlStatement := `
		SELECT (id) FROM users
		WHERE email = ($1)
	`
	db := postgresql.GetDB()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Something wrong witg request. Please check it :)t")
	}
	cart.Product_id = id

	if err := c.Bind(&cart.Count); err != nil {
		log.Infof("Cannot bind the request: %v \n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Something wrong witg request. Please check it :)")
	}

	claims, err := u.GetClaims(c.Request())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, 0)
	}
	email := claims["email"]

	err = db.QueryRow(context.Background(), sqlStatement, email).Scan(&cart.User_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal server error")
	}

	id, err = crt.AddProductToCart(cart)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal server error")
	}
	return c.JSON(http.StatusOK, id)
}
