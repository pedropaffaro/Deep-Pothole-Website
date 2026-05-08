package handlers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pedropaffaro/deep-pothole-backend/internal/database"
	"github.com/pedropaffaro/deep-pothole-backend/internal/models"

	"github.com/gofiber/fiber/v2"
)

func CreateComplaint(c *fiber.Ctx) error {
	// Capturar campos de texto do formulário
	city := c.FormValue("cidade")
	street := c.FormValue("rua")
	latStr := c.FormValue("latitude")
	lngStr := c.FormValue("longitude")

	// converter strings para float64
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	file, err := c.FormFile("foto")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Foto é obrigatória"})
	}

	// Criar um nome único para o arquivo para não sobrescrever
	fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join("./uploads", fileName)

	// Salvar o arquivo fisicamente na pasta /uploads
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao salvar imagem localmente"})
	}

	// 3. Criar o objeto para o Banco de Dados
	novaDenuncia := models.Complaint{
		City:    city,
		Street:       street,
		Latitude:  lat,
		Longitude: lng,
		PhotoURL:   filePath, // Guardamos o caminho do arquivo no .db
	}

	// 4. O "Pulo do Gato": Mandar para o arquivo .db
	// O GORM faz todo o trabalho de SQL por baixo dos panos
	result := database.DB.Create(&novaDenuncia) 
	
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao inserir no banco de dados"})
	}

	// Retornar o objeto criado (já com o ID que o SQLite gerou)
	return c.Status(201).JSON(novaDenuncia)
}