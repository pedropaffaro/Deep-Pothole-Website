package database

import (
	"fmt"
	"log"

	"github.com/pedropaffaro/deep-pothole-backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error

	DB, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	
	if err != nil {
		log.Fatal("Falha ao conectar ao banco SQLite:", err)
	}

	fmt.Println("Banco conectado.")

	err = DB.AutoMigrate(&models.Complaint{})
	if err != nil {
		log.Fatal("Erro ao executar migration:", err)
	}
}