package handlers

import (
	"strconv"

	cf "github.com/pedropaffaro/deep-pothole-backend/internal/cloudflare"

	"github.com/gofiber/fiber/v2"
)

func CreateComplaint(c *fiber.Ctx) error {
	city := c.FormValue("cidade")
	street := c.FormValue("rua")
	latStr := c.FormValue("latitude")
	lngStr := c.FormValue("longitude")

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	file, err := c.FormFile("foto")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Foto é obrigatória"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao abrir imagem"})
	}
	defer src.Close()

	// fileData, err := io.ReadAll(src)
	// if err != nil {
	// 	return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler imagem"})
	// }

	cfg := cf.LoadConfig()

	// objectKey := fmt.Sprintf("photos/%d-%s", time.Now().Unix(), file.Filename)

	// photoURL, err := cf.PostToR2(cfg, objectKey, fileData, file.Filename)
	// if err != nil {
	// 	return c.Status(500).JSON(fiber.Map{"error": "Erro ao enviar imagem do buraco: " + err.Error()})
	// }
	photoURL := "https://link-da-imagem-no-r2.com/"
	
	sql := `INSERT INTO complaints (city, street, latitude, longitude, photo_url)
	        VALUES (?, ?, ?, ?, ?)`

	lastID, err := cf.PostToD1(cfg, sql, []any{city, street, lat, lng, photoURL})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao salvar registro de buraco: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":        lastID,
		"city":      city,
		"street":    street,
		"latitude":  lat,
		"longitude": lng,
		"photo_url": photoURL,
	})
}