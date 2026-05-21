package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/routex/routex/internal/model"
	"github.com/routex/routex/internal/xray"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
	xc *xray.Client
}

func NewUserService(db *gorm.DB, xc *xray.Client) *UserService {
	return &UserService{db: db, xc: xc}
}

type CreateUserInput struct {
	Username     string    `json:"username"      binding:"required"`
	TrafficLimit int64     `json:"traffic_limit"`
	ExpireTime   time.Time `json:"expire_time"`
}

// Create 先在 DB 建用户拿到 ID，再调 Xray AddUser。
// Xray 失败时回滚 DB，保证两边一致。
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*model.User, error) {
	u := &model.User{
		Username:     in.Username,
		UUID:         uuid.NewString(),
		Status:       1,
		TrafficLimit: in.TrafficLimit,
		ExpireTime:   in.ExpireTime,
	}
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return nil, err
	}
	if err := s.xc.AddUser(ctx, s.specOf(u)); err != nil {
		// 回滚：删 DB 记录，否则下次再创建会冲突
		_ = s.db.WithContext(ctx).Delete(u).Error
		return nil, fmt.Errorf("xray add user: %w", err)
	}
	return u, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	var users []model.User
	return users, s.db.WithContext(ctx).Order("id desc").Find(&users).Error
}

// Delete DB + Xray 双删。Xray 报 not-found 时容忍（可能已被禁用清掉）。
func (s *UserService) Delete(ctx context.Context, id int64) error {
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return err
	}
	if err := s.xc.RemoveUser(ctx, xray.EmailOf(id)); err != nil && !isNotFound(err) {
		return fmt.Errorf("xray remove: %w", err)
	}
	return s.db.WithContext(ctx).Delete(&u).Error
}

// Disable 把用户置为禁用并从 Xray 移除（连接断开），但保留 DB 行。
func (s *UserService) Disable(ctx context.Context, id int64) error {
	u, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if u.Status == 0 {
		return nil
	}
	if err := s.xc.RemoveUser(ctx, xray.EmailOf(id)); err != nil && !isNotFound(err) {
		return fmt.Errorf("xray remove: %w", err)
	}
	return s.db.WithContext(ctx).Model(u).Update("status", 0).Error
}

// Enable 重新把用户加回 Xray 并置 status=1。
func (s *UserService) Enable(ctx context.Context, id int64) error {
	u, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if u.Status == 1 {
		return nil
	}
	if err := s.xc.AddUser(ctx, s.specOf(u)); err != nil {
		return fmt.Errorf("xray add: %w", err)
	}
	return s.db.WithContext(ctx).Model(u).Update("status", 1).Error
}

func (s *UserService) specOf(u *model.User) xray.UserSpec {
	return xray.UserSpec{
		Email:    xray.EmailOf(u.ID),
		UUID:     u.UUID,
		Flow:     "xtls-rprx-vision",
		Protocol: xray.ProtocolVLESS,
	}
}

// SyncAll 启动时把 DB 里 status=1 的用户全部 push 到 Xray（容忍 already exists）。
// 用于 Xray 容器重启后丢失运行态用户的恢复。
func (s *UserService) SyncAll(ctx context.Context) (added, skipped int, err error) {
	var users []model.User
	if err = s.db.WithContext(ctx).Where("status = ?", 1).Find(&users).Error; err != nil {
		return
	}
	for _, u := range users {
		spec := s.specOf(&u)
		if err2 := s.xc.AddUser(ctx, spec); err2 != nil {
			if isAlreadyExists(err2) {
				skipped++
				continue
			}
			err = err2
			return
		}
		added++
	}
	return
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func isAlreadyExists(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate"))
}

// ErrUserNotFound 业务层封装的 not-found 哨兵。
var ErrUserNotFound = errors.New("user not found")
