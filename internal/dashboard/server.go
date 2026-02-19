package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
)

//go:embed index.html
var content embed.FS

type BackupFile struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Time string `json:"time"`
}

func StartServer(port int) error {
	backupDir := "./backups"
	dirExists := true
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		dirExists = false
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/backups", func(w http.ResponseWriter, r *http.Request) {
		if !dirExists {
			json.NewEncoder(w).Encode([]BackupFile{})
			return
		}

		files, _ := os.ReadDir(backupDir)
		type fileInfo struct {
			name string
			size int64
			time time.Time
		}
		var tempFiles []fileInfo

		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".zip" {
				info, _ := f.Info()
				tempFiles = append(tempFiles, fileInfo{
					name: f.Name(),
					size: info.Size(),
					time: info.ModTime(),
				})
			}
		}

		sort.Slice(tempFiles, func(i, j int) bool {
			return tempFiles[i].time.After(tempFiles[j].time)
		})

		backupList := []BackupFile{}
		for _, f := range tempFiles {
			backupList = append(backupList, BackupFile{
				Name: f.name,
				Size: formatSize(f.size),
				Time: f.time.Format("2006-01-02 15:04:05"),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(backupList)
	})

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		files, _ := os.ReadDir(backupDir)
		var totalSize int64
		var fileCount int
		var lastBackup string

		type fileTemp struct {
			name string
			time time.Time
		}
		var tempFiles []fileTemp

		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".zip" {
				info, _ := f.Info()
				totalSize += info.Size()
				fileCount++
				tempFiles = append(tempFiles, fileTemp{name: f.Name(), time: info.ModTime()})
			}
		}

		if fileCount > 0 {
			sort.Slice(tempFiles, func(i, j int) bool {
				return tempFiles[i].time.After(tempFiles[j].time)
			})
			lastBackup = tempFiles[0].name
		} else {
			lastBackup = "-"
		}

		stats := map[string]interface{}{
			"total_files": fileCount,
			"total_size":  formatSize(totalSize),
			"last_backup": lastBackup,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/api/delete/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fileName := filepath.Base(r.URL.Path)
		filePath := filepath.Join(backupDir, fileName)

		if err := os.Remove(filePath); err != nil {
			http.Error(w, "Failed to delete file", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "File %s deleted", fileName)
	})

	mux.Handle("/api/download/", http.StripPrefix("/api/download/", http.FileServer(http.Dir(backupDir))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index, err := content.ReadFile("index.html")
		if err != nil {
			http.Error(w, "File index.html tidak ditemukan dalam binary", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(index)
	})

	fmt.Printf("\n%s🚀 Dashboard Started!%s\n", constant.ColorGreen, constant.ColorReset)
	fmt.Printf("%s🌐 Local  : %shttp://localhost:%d%s\n",
		constant.ColorCyan, constant.ColorWhite, port, constant.ColorReset)

	if !dirExists {
		fmt.Printf("%s⚠️ Warning: Directory %s%s%s does not exist!%s\n",
			constant.ColorYellow, constant.ColorWhite, backupDir, constant.ColorYellow, constant.ColorReset)
	}

	fmt.Printf("%s💡 Press %sCtrl+C%s %sto stop the dashboard server.%s\n",
		constant.ColorGray, constant.ColorYellow, constant.ColorGray, constant.ColorGray, constant.ColorReset)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGT"[exp])
}
