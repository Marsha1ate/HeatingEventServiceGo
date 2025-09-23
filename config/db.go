package config

import (
	"HeatingEventServiceGo/models"
	"fmt"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	dbName := GetEnv("DB_NAME", "messages")
	dbUser := GetEnv("DB_USER", "postgres")
	dbPassword := GetEnv("DB_PASSWORD", "postgres")
	dbHost := GetEnv("DB_HOST", "postgres")
	dbPortStr := GetEnv("PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		dbPort = 5432
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("Failed to connect to database")
	}
	if err := DB.AutoMigrate(&models.Message{}); err != nil {
		panic("AutoMigrate failed: " + err.Error())
	}

	// Установка дефолтов для колонок (если таблица существует)
	DB.Exec("ALTER TABLE IF EXISTS messages ALTER COLUMN server_time SET DEFAULT CURRENT_TIMESTAMP;")
	DB.Exec("ALTER TABLE IF EXISTS messages ALTER COLUMN source_time SET DEFAULT CURRENT_TIMESTAMP;")

	// Фикс старых записей: установить NOW() где zero
	result := DB.Exec("UPDATE messages SET server_time = CURRENT_TIMESTAMP WHERE server_time = '0001-01-01 00:00:00'::timestamp;")
	if result.Error != nil {
		fmt.Println("Ошибка обновления server_time:", result.Error)
	} else {
		fmt.Println("Обновлено записей server_time:", result.RowsAffected)
	}
	result = DB.Exec("UPDATE messages SET source_time = CURRENT_TIMESTAMP WHERE source_time = '0001-01-01 00:00:00'::timestamp;")
	if result.Error != nil {
		fmt.Println("Ошибка обновления source_time:", result.Error)
	} else {
		fmt.Println("Обновлено записей source_time:", result.RowsAffected)
	}
	fmt.Println("DATABASE CONNECTED")
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
