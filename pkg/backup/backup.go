package backup

import (
	"HibiscusIM/pkg/config"
	"HibiscusIM/pkg/logger"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// StartBackupScheduler 启动备份调度器
func StartBackupScheduler() {
	c := cron.New()

	// 使用配置中的 Cron 表达式
	schedule := config.GlobalConfig.BackupSchedule

	// 添加定时任务
	c.AddFunc(schedule, func() {
		err := ExecuteBackup()
		if err != nil {
			logger.Warn("Backup failed: %v", zap.Error(err))
		} else {
			logger.Info("Backup completed successfully")
		}
	})

	// 启动调度器
	c.Start()
}

// ExecuteBackup 根据配置执行数据库备份
func ExecuteBackup() error {
	switch config.GlobalConfig.DBDriver {
	case "sqlite":
		// 执行 SQLite 备份
		dst := filepath.Join(config.GlobalConfig.BackupPath, fmt.Sprintf("sys_backup_%s.db", time.Now().Format("20060102_150405")))
		return BackupSQLiteDatabase(config.GlobalConfig.DSN, dst)
	case "mysql":
		// 执行 MySQL 备份
		dst := filepath.Join(config.GlobalConfig.BackupPath, fmt.Sprintf("sys_backup_%s.sql", time.Now().Format("20060102_150405")))
		return BackupMySQLDatabase(config.GlobalConfig.DSN, dst)
	default:
		return fmt.Errorf("unsupported DB_DRIVER: %s", config.GlobalConfig.DBDriver)
	}
}

// BackupSQLiteDatabase 执行 SQLite 数据库的备份
func BackupSQLiteDatabase(src string, dst string) error {
	// 确保目标路径存在
	backupDir := filepath.Dir(dst)
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		err := os.MkdirAll(backupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create backup directory: %v", err)
		}
	}

	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %v", err)
	}
	defer sourceFile.Close()

	// 创建备份文件
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err)
	}
	defer destFile.Close()

	// 拷贝数据
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying data: %v", err)
	}

	log.Printf("SQLite database backup completed: %s", dst)
	return nil
}

// BackupMySQLDatabase 执行 MySQL 数据库的备份
func BackupMySQLDatabase(dsn, dst string) error {
	// 确保目标路径存在
	backupDir := filepath.Dir(dst)
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		err := os.MkdirAll(backupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create backup directory: %v", err)
		}
	}

	// 使用 mysqldump 执行备份
	cmd := exec.Command("mysqldump", dsn, ">", dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to backup MySQL database: %v", err)
	}

	log.Printf("MySQL database backup completed: %s", dst)
	return nil
}
