package main

import (
	"fmt"
	"net/http"
	"vinylShop/config"
	h "vinylShop/internal/handlers"

	"github.com/golang-jwt/jwt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

var (
	cfg config.Properties
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v\n", err)
	}
}

func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		jwtToken := c.Request().Header["X-Access-Token"][0]
		claims := jwt.MapClaims{}

		_, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtTokenSecret), nil
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Unable to parse token")
		}
		if !claims["authorized"].(bool) {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}
		return next(c)
	}
}

func main() {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	// Users handlers
	e.POST("/user", h.CreateUser)
	e.POST("/auth", h.AuthnUser)
	e.PUT("/user", h.UpdateUser)
	e.POST("/refresh", h.RefreshToken)

	// Products handlers
	e.POST("/product", h.CreateProduct, adminMiddleware)
	e.DELETE("/product/:id", h.DeleteProduct, adminMiddleware)
	e.GET("/product/:id", h.GetProduct)
	e.PUT("/product", h.UpdateProduct, adminMiddleware)

	// Orders handlers: add, get, delete
	e.POST("/cart/:id", h.AddProductToCart)
	

	e.Logger.Infof("Listening on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}