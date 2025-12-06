package test

import (
	"bytes"
	"encoding/json"
	"gouas/database"
	"gouas/route"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// SetupTestApp mempersiapkan aplikasi fiber untuk testing
func SetupTestApp() *fiber.App {
	// Load .env dari folder parent (karena file ini ada di dalam folder gouas/test/)
	// Jika gagal load, asumsinya env vars sudah ada di system environment
	godotenv.Load("../.env") 
	
	// Pastikan koneksi DB terhubung
	database.ConnectPostgres()
	
	app := fiber.New()
	
	// Load Route agar endpoint /api/v1/auth/login tersedia
	route.SetupRoutes(app)
	
	return app
}

func TestLogin_Success(t *testing.T) {
	app := SetupTestApp()

	// Data login (Pastikan user "admin" sudah ada di database dari proses Seeder di Tahap 2)
	payload := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(payload)

	// Buat Request HTTP POST ke endpoint login
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Eksekusi Request
	resp, err := app.Test(req)

	// Validasi Hasil
	assert.Nil(t, err) // Pastikan tidak ada error internal
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Pastikan status code 200 OK
}

func TestLogin_Fail(t *testing.T) {
	app := SetupTestApp()

	// Data login salah
	payload := map[string]string{
		"username": "admin",
		"password": "salahpassword",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Validasi Hasil
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) // Pastikan status code 401 Unauthorized
}