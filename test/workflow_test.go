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

// Helper function getAdminToken diasumsikan ada di file user_test.go (package test yang sama)

func TestVerifyAchievement_Success(t *testing.T) {
	// 1. Setup Environment
	godotenv.Load("../.env")
	database.ConnectPostgres()
	database.ConnectMongo()
	app := fiber.New()
	route.SetupRoutes(app)

	// 2. Login Admin (Kita gunakan Admin sebagai simulasi user serba bisa untuk test ini)
	adminToken := getAdminToken(t, app)

	// 3. Setup Data: Buat Admin menjadi Dosen sekaligus Mahasiswa Bimbingannya sendiri
	// Tujuannya agar kita bisa test create (sebagai mhs) dan verify (sebagai dosen) dengan satu token
	
	var adminUser model.User
	if err := database.DB.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		t.Fatal("User admin tidak ditemukan. Pastikan seeder sudah dijalankan.")
	}
	
	// A. Setup Profile Dosen
	lecturer := model.Lecturer{
		UserID:     adminUser.ID,
		LecturerID: "NIP12345",
		Department: "IT",
	}
	// Upsert Lecturer (Buat baru atau update jika ada)
	database.DB.Where(model.Lecturer{UserID: adminUser.ID}).Assign(lecturer).FirstOrCreate(&lecturer)

	// B. Setup Profile Mahasiswa (Link ke Dosen di atas)
	student := model.Student{
		UserID:       adminUser.ID,
		NIM:          "NIMTEST123",
		ProgramStudy: "IT",
		AcademicYear: "2025",
		AdvisorID:    &lecturer.ID, // PENTING: Assign diri sendiri sebagai advisor
	}
	database.DB.Where(model.Student{UserID: adminUser.ID}).Assign(student).FirstOrCreate(&student)

	// 4. Create Achievement (Action: Create Draft)
	payload := map[string]interface{}{
		"achievement_type": "competition",
		"title":            "Lomba Test Workflow",
		"description":      "Deskripsi Lomba untuk Test",
		"details":          map[string]interface{}{"rank": 1},
	}
	body, _ := json.Marshal(payload)
	
	reqCreate := httptest.NewRequest("POST", "/api/v1/achievements", bytes.NewReader(body))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+adminToken)
	
	respCreate, err := app.Test(reqCreate)
	assert.Nil(t, err)

	// --- ERROR HANDLING & PARSING RESPONSE (Agar tidak Panic) ---
	
	// Cek Status Code Create
	if respCreate.StatusCode != 200 {
		var errBody map[string]interface{}
		json.NewDecoder(respCreate.Body).Decode(&errBody)
		t.Fatalf("Gagal Create Achievement. Status: %d. Body: %v", respCreate.StatusCode, errBody)
	}

	// Decode JSON Response
	var resCreate map[string]interface{}
	if err := json.NewDecoder(respCreate.Body).Decode(&resCreate); err != nil {
		t.Fatal("Gagal decode response body:", err)
	}

	// Ambil Data Object
	data, ok := resCreate["data"].(map[string]interface{})
	if !ok || data == nil {
		t.Fatal("Response 'data' kosong atau format salah:", resCreate)
	}

	// Ambil ID (Pastikan Model AchievementReference sudah ada tag `json:"id"`)
	achID, ok := data["id"].(string)
	if !ok {
		t.Fatal("Field 'id' tidak ditemukan di response data. Cek struct model AchievementReference.")
	}

	// 5. Verify Achievement (Action: Dosen Verify)
	reqVerify := httptest.NewRequest("POST", "/api/v1/achievements/"+achID+"/verify", nil)
	reqVerify.Header.Set("Authorization", "Bearer "+adminToken)
	
	respVerify, err := app.Test(reqVerify)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, respVerify.StatusCode, "Gagal verifikasi prestasi")
}