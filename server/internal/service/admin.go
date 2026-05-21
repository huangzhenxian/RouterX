package service

import (
	"context"
	"errors"
	"time"

	"github.com/routex/routex/internal/auth"
	"github.com/routex/routex/internal/model"
	"github.com/routex/routex/internal/util"
	"gorm.io/gorm"
)

type AdminService struct {
	db        *gorm.DB
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAdminService(db *gorm.DB, jwtSecret string) *AdminService {
	return &AdminService{
		db:        db,
		jwtSecret: jwtSecret,
		tokenTTL:  12 * time.Hour,
	}
}

var ErrLoginFailed = errors.New("invalid username or password")

type LoginResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Admin     *model.Admin `json:"admin"`
}

func (s *AdminService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var a model.Admin
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&a).Error
	if err != nil || !util.VerifyPassword(a.PasswordHash, password) {
		return nil, ErrLoginFailed
	}
	token, err := auth.Sign(s.jwtSecret, a.ID, a.Username, a.Role, s.tokenTTL)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token:     token,
		ExpiresAt: time.Now().Add(s.tokenTTL),
		Admin:     &a,
	}, nil
}
