package conexoes

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConectarPostgreSQL(database string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s  dbname=%s user=%s password=%s",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), database,
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}
