package backup

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
	"github.com/GeraAnggaraPutra/go-backup/internal/notification"
	"github.com/GeraAnggaraPutra/go-backup/internal/storage"
)

// checks if the string matches the YYYYMMDD_HHMMSS pattern
func isDefaultTimestamp(s string) bool {
	if len(s) < 15 {
		return false
	}
	for i, r := range s {
		if i == 8 { // The underscore position
			if r != '_' {
				return false
			}
			continue
		}

		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func generateTimestampedOutput(originalOutput string, now time.Time) string {
	timestamp := now.Format("20060102_150405")

	// Separate folder path and file name
	pathSeparatorIdx := strings.LastIndex(originalOutput, "/")
	folderPath := ""
	fileName := originalOutput

	if pathSeparatorIdx != -1 {
		folderPath = originalOutput[:pathSeparatorIdx+1] // example: "backups/"
		fileName = originalOutput[pathSeparatorIdx+1:]   // example: "backup.zip"
	}

	// If the file name starts with a default timestamp pattern, remove it before adding the new timestamp
	if len(fileName) >= 16 && isDefaultTimestamp(fileName[:15]) {
		parts := strings.SplitN(fileName, "_", 3)
		if len(parts) >= 3 {
			fileName = parts[2]
		}
	}

	// Construct the new output path with the current timestamp
	return fmt.Sprintf("%s%s_%s", folderPath, timestamp, fileName)
}

func uploadToGCS(ctx context.Context, cfg Config) {
	if !cfg.ENVConfig.GCSEnabled {
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

func sendTelegramNotification(cfg Config, errBackup error) {
	if !cfg.ENVConfig.TelegramEnabled {
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

	if errNotif := notification.SendToTelegram(cfg.ENVConfig.TelegramToken, cfg.ENVConfig.TelegramChatID, filePath, msg); errNotif != nil {
		fmt.Printf("\n%s[Notification] Failed to send Telegram: %v%s\n", constant.ColorRed, errNotif, constant.ColorReset)
	} else {
		fmt.Printf("%s[Notification] Telegram notification sent successfully!%s\n", constant.ColorGreen, constant.ColorReset)
	}
}
