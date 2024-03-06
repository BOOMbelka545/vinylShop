package postgresql

import (
	"context"
	"fmt"
	"vinylShop/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/gommon/log"
)

type (
	User struct {
		Id       int    `json:"id,omitempty"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password,omitempty" validate:"min=8,max=300,required"`
		IsAdmin  bool   `json:"isadmin,omitempty"`
	}
	
	Product struct {
		Id          int    `json:"id,omitempty"`
		Name        string `json:"name" validate:"required"`
		Cost        int    `json:"cost,omitempty" validate:"min=0,required"`
		ArtistName  string `json:"artistName,omitempty"`
	}

	Cart struct {
		Id           int  `json:"id,omitempty"`
		User_id      int  `json:"user_id"`
		Product_id   int  `json:"product_id"`
		Count        int  `json:"count"`
	}
)

var (
	conn *pgx.Conn
	cfg   config.Properties
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v\n", err)
	}

	conn = connToDB(cfg)

	// ! Maybe it doesn't work. I didn't check it but it looks good
	// * Check migration file
	sqlStatement := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			isadmin bool
		)
	`
	_, err := conn.Exec(context.Background(), sqlStatement)
	if err != nil {
		log.Fatalf("Table users connot be created: %v\n", err)
	}

	sqlStatement = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			cost integer NOT NULL,
			artistname VARCHAR(255) NOT NULL
		)
	`
	_, err = conn.Exec(context.Background(), sqlStatement)
	if err != nil {
		log.Fatalf("Table products connot be created: %v\n", err)
	}
}

func connToDB(cfg config.Properties) *pgx.Conn {
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatalf("Unable to read the env: %v\n", err)
	}

	// Connect to db
	// URLexample = 'postgres://Login:Password@Host:Port/dbName'
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DBLog, cfg.DBPas, cfg.DBHost, cfg.DBPort, cfg.DBName)
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to db: %v\n %v", err, dbURL)
	}

	// defer conn.Close(context.Background())

	return conn
}

func GetDB() *pgx.Conn {
	return conn
}