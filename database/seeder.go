package database

import (
	"gouas/app/model"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func SeedDatabase() {
	db := DB

	// 1. Seed Roles
	roles := []model.Role{
		{Name: "Admin", Description: "Administrator Sistem"},
		{Name: "Mahasiswa", Description: "Pelapor Prestasi"},
		{Name: "Dosen Wali", Description: "Verifikator Prestasi"},
	}

	for _, r := range roles {
		if err := db.FirstOrCreate(&model.Role{}, model.Role{Name: r.Name}).Error; err != nil {
			log.Printf("Failed to seed role %s: %v", r.Name, err)
		}
	}

	// 2. Seed Permissions
	permissions := []model.Permission{
		{Name: "achievement:create", Resource: "achievement", Action: "create"},
		{Name: "achievement:read", Resource: "achievement", Action: "read"},
		{Name: "achievement:verify", Resource: "achievement", Action: "verify"},
		{Name: "user:manage", Resource: "user", Action: "manage"},
	}

	for _, p := range permissions {
		if err := db.FirstOrCreate(&model.Permission{}, model.Permission{Name: p.Name}).Error; err != nil {
			log.Printf("Failed to seed permission %s: %v", p.Name, err)
		}
	}

	// 3. Seed Super Admin User
	var adminRole model.Role
	db.Where("name = ?", "Admin").First(&adminRole)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	adminUser := model.User{
		Username:     "admin",
		Email:        "admin@gouas.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Super Administrator",
		RoleID:       adminRole.ID,
		IsActive:     true,
	}

	if err := db.Where("username = ?", adminUser.Username).FirstOrCreate(&adminUser).Error; err != nil {
		log.Printf("Failed to seed admin user: %v", err)
	}

	// ==========================================
	// 4. ASSIGN PERMISSIONS TO ADMIN (FIX 403)
	// ==========================================
	var allPermissions []model.Permission
	db.Find(&allPermissions) // Ambil semua permission yang ada

	// Masukkan semua permission ke role Admin lewat tabel pivot role_permissions
	if err := db.Model(&adminRole).Association("Permissions").Replace(allPermissions); err != nil {
		log.Printf("Failed to assign permissions to admin: %v", err)
	} else {
		log.Println("✅ Permissions assigned to Admin Role!")
	}

	log.Println("✅ Database Seeding Completed!")
}