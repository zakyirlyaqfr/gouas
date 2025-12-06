package test

import (
	"bytes"
	"encoding/json"
	"gouas/app/model"
	"gouas/database"
	"gouas/route"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// Helper untuk login sebagai admin dan dapat token
func getAdminToken(t *testing.T, app *fiber.App) string {
	payload := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatal("Request login error:", err)
	}

	if resp.StatusCode != 200 {
		t.Fatal("Gagal login admin. Pastikan user 'admin' ada di DB. Status Code:", resp.StatusCode)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal("Gagal decode response login")
	}

	// Cek struktur JSON response standar kita { "data": { "token": "..." } }
	data, ok := res["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Format response data login salah")
	}

	token, ok := data["token"].(string)
	if !ok {
		t.Fatal("Token tidak ditemukan dalam response")
	}

	return token
}

func TestSetupStudent_Success(t *testing.T) {
	// Setup App
	godotenv.Load("../.env")
	database.ConnectPostgres()
	app := fiber.New()
	route.SetupRoutes(app)

	// 1. Get Admin Token (Dengan Error Handling)
	token := getAdminToken(t, app)

	// 2. Ambil user admin dari DB untuk test ID-nya
	var adminUser model.User
	if err := database.DB.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		t.Fatal("User admin tidak ditemukan di database untuk test setup student")
	}

	// 3. Payload Setup Student
	payload := map[string]string{
		"user_id":       adminUser.ID.String(),
		"nim":           "123456789",
		"program_study": "Informatika",
		"academic_year": "2025",
	}
	body, _ := json.Marshal(payload)

	// 4. Request dengan Auth Header
	req := httptest.NewRequest("POST", "/api/v1/users/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}