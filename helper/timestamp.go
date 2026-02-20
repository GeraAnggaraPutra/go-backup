package helper

import (
	"fmt"
	"strings"
	"time"
)

// checks if the string matches the YYYYMMDD_HHMMSS pattern
func IsDefaultTimestamp(s string) bool {
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

func GenerateTimestampedOutput(originalOutput string, now time.Time) string {
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
	if len(fileName) >= 16 && IsDefaultTimestamp(fileName[:15]) {
		parts := strings.SplitN(fileName, "_", 3)
		if len(parts) >= 3 {
			fileName = parts[2]
		}
	}

	// Construct the new output path with the current timestamp
	return fmt.Sprintf("%s%s_%s", folderPath, timestamp, fileName)
}
