package auth

import (
	"github.com/routex/routex/internal/model"
	"github.com/routex/routex/internal/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnsureDefaultAdmin 首启时若无任何管理员，自动创建 admin/<随机密码>，
// 并把明文密码醒目地打到日志里（前缀 BOOTSTRAP_ADMIN，方便 grep 捞）。
//
// 后续启动若 admins 表已有记录，本函数空操作；不会覆盖现有密码。
func EnsureDefaultAdmin(db *gorm.DB, log *zap.Logger) error {
	var count int64
	if err := db.Model(&model.Admin{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	plain := util.RandPassword(20)
	hash, err := util.HashPassword(plain)
	if err != nil {
		return err
	}
	admin := &model.Admin{
		Username:     "admin",
		PasswordHash: hash,
		Role:         "super",
	}
	if err := db.Create(admin).Error; err != nil {
		return err
	}

	// 用 Warn 级别 + 高亮分隔线，stdout/日志文件里非常显眼。
	// 一行 key=value 形式方便 `grep BOOTSTRAP_ADMIN logs/server.log` 直接捞。
	log.Warn("=============================================================")
	log.Warn("BOOTSTRAP_ADMIN created (saved this line — password is plaintext, won't show again)")
	log.Warn("BOOTSTRAP_ADMIN username=admin password=" + plain)
	log.Warn("=============================================================")
	return nil
}
