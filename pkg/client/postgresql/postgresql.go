package postgresql

import (
	"context"
	"fmt"
	"vinylShop/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/gommon/log"
)

// var (
// 	cfg config.DBConfig
// )

// Structure users table 
type User struct {
	Id       int    `json:"id,omitempty"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"min=8,max=300,required"`
	IsAdmin  bool   `json:"isadmin,omitempty"`
}

// Structure products table
type Product struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name" validate:"required"`
	Cost        int    `json:"cost,omitempty" validate:"min=0,required"`
	ArtistName  string `json:"artistName,omitempty"`
}

var (
	conn *pgx.Conn
	cfg   config.Properties
)

func init() {
	// Read env for echo and DB
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Config cannot be read: %v\n", err)
	}

	// Connect to db
	conn = connToDB(cfg)
}

// Connect to postgres DB
func connToDB(cfg config.Properties) *pgx.Conn {
	// Read env to cfg
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