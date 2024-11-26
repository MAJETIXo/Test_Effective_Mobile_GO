package migrations

import (
	"log"
	"server/config"
	"server/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MigrateDB() {
  dsn := config.GetDBConfig()
  db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
  if err != nil {
    log.Fatalf("Error connecting to the database: %v", err)
  }
  err = db.AutoMigrate(&models.Group{}, &models.Song{})
  if err != nil {
    log.Fatalf("Error during migration: %v", err)
  }
  log.Println("Migration completed successfully!")
}