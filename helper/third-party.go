package helper

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
	"github.com/GeraAnggaraPutra/go-backup/internal/notification"
	"github.com/GeraAnggaraPutra/go-backup/internal/storage"
)

type UploadGCSRequest struct {
	GCSEnabled bool
	Output     string
}

func UploadToGCS(ctx context.Context, cfg UploadGCSRequest) {
	if !cfg.GCSEnabled {
		return
	}

	driveStorage, err := storage.NewStorage()
	if err != nil {
		fmt.Printf("%s[Google Cloud Storage] Failed to initialize GCS client: %v%s\n", constant.ColorRed, err, constant.ColorReset)
		return
	}

	if err := driveStorage.UploadToGCS(ctx, cfg.Output); err != nil {
		fmt.Printf("%s[Google Cloud Storage] Upload failed: %v%s\n", constant.ColorRed, err, constant.ColorReset)
	} else {
		fmt.Printf("%s[Google Cloud Storage] Upload successful!%s\n", constant.ColorGreen, constant.ColorReset)
	}
}

type UploadTelegramRequest struct {
	TelegramEnabled bool
	Output          string
	DBName          string
	TelegramToken   string
	TelegramChatID  string
}

func SendTelegramNotification(cfg UploadTelegramRequest, errBackup error) {
	if !cfg.TelegramEnabled {
		return
	}

	finishTime := time.Now().Format("2006-01-02 15:04:05")

	var msg string
	var filePath string

	if errBackup != nil {
		msg = fmt.Sprintf(constant.TelegramErrorTemplate, cfg.DBName, errBackup.Error(), finishTime)
	} else {
		fileNameOnly := filepath.Base(cfg.Output)
		msg = fmt.Sprintf(constant.TelegramMessageTemplate, cfg.DBName, fileNameOnly, finishTime)
		filePath = cfg.Output
	}

	if errNotif := notification.SendToTelegram(cfg.TelegramToken, cfg.TelegramChatID, filePath, msg); errNotif != nil {
		fmt.Printf("\n%s[Notification] Failed to send Telegram: %v%s\n", constant.ColorRed, errNotif, constant.ColorReset)
	} else {
		fmt.Printf("%s[Notification] Telegram notification sent successfully!%s\n", constant.ColorGreen, constant.ColorReset)
	}
}
