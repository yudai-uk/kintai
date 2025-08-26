package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"
    "github.com/labstack/echo/v4"
    echomw "github.com/labstack/echo/v4/middleware"
    "github.com/yudai-uk/backend/models"
    "github.com/yudai-uk/backend/routes"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set in the .env file")
	}

	if os.Getenv("SUPABASE_JWT_SECRET") == "" {
		log.Fatal("SUPABASE_JWT_SECRET is not set in the .env file")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Attendance{},
		&models.Leave{},
		&models.Schedule{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	e := echo.New()

    e.Use(echomw.Logger())
    e.Use(echomw.Recover())
    e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
        AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{"Content-Type", "Authorization"},
    }))

	routes.SetupRoutes(e, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(":" + port))
}
