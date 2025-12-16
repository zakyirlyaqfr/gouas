package repository

import (
	"gouas/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository interface {
	FindByUsername(username string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	// [BARU] Update whitelist token ID
	UpdateTokenIDs(userID uuid.UUID, accessID *uuid.UUID, refreshID *uuid.UUID) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role.Permissions").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// [BARU] Implementasi Update Token IDs
func (r *authRepository) UpdateTokenIDs(userID uuid.UUID, accessID *uuid.UUID, refreshID *uuid.UUID) error {
	updates := map[string]interface{}{}
	
	// Selalu update Access Token ID (bisa UUID baru atau nil saat logout)
	updates["current_access_token_id"] = accessID

	// Refresh Token ID diupdate hanya jika parameter tidak nil 
	// (Saat refresh token, kita biarkan refresh token ID tetap sama, kecuali logout)
	if refreshID != nil {
		updates["current_refresh_token_id"] = refreshID
	} else if accessID == nil {
		// Jika AccessID nil (Logout), maka RefreshID juga harus dihapus (nil)
		updates["current_refresh_token_id"] = nil
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}