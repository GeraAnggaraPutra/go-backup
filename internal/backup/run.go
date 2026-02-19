package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GeraAnggaraPutra/go-backup/internal/config"
	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
	"github.com/GeraAnggaraPutra/go-backup/internal/database"
	"github.com/GeraAnggaraPutra/go-backup/internal/notification"
	"github.com/GeraAnggaraPutra/go-backup/internal/zipper"

	"github.com/robfig/cron/v3"
)

type Config struct {
	DBType    string
	Host      string
	Port      int
	Username  string
	Password  string
	DBName    string
	Output    string
	Tables    []string
	Schedule  string
	ENVConfig config.Config
}

const defaultWorkerCount = 4

func Run(cfg Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Printf("\r%s%s[System] Interrupt signal received. Cleaning up...%s\n",
			constant.ClearLine, constant.ColorYellow, constant.ColorReset)
		cancel()
	}()

	processAction := func(currentCfg Config) error {
		err := executeBackup(ctx, currentCfg)

		finishTime := time.Now().Format("2006-01-02 15:04:05")
		hasTele := currentCfg.ENVConfig.TelegramToken != "" && currentCfg.ENVConfig.TelegramChatID != ""

		if err != nil {
			if hasTele {
				msg := fmt.Sprintf(constant.TelegramErrorTemplate,
					currentCfg.DBName,
					err.Error(),
					finishTime,
				)

				if errNotif := notification.SendToTelegram(currentCfg.ENVConfig.TelegramToken, currentCfg.ENVConfig.TelegramChatID, "", msg); errNotif != nil {
					fmt.Printf("\n%s[Notification] Failed to send Telegram error notification: %v%s\n", constant.ColorRed, errNotif, constant.ColorReset)
				} else {
					fmt.Printf("%s[Notification] Error notification sent to Telegram successfully!%s\n", constant.ColorGreen, constant.ColorReset)
				}
			}

			return err
		}

		if hasTele {
			fileNameOnly := filepath.Base(currentCfg.Output)

			msg := fmt.Sprintf(constant.TelegramMessageTemplate,
				currentCfg.DBName,
				fileNameOnly,
				finishTime,
			)

			if errNotif := notification.SendToTelegram(currentCfg.ENVConfig.TelegramToken, currentCfg.ENVConfig.TelegramChatID, currentCfg.Output, msg); errNotif != nil {
				fmt.Printf("\n%s[Notification] Failed to send Telegram: %v%s\n", constant.ColorRed, errNotif, constant.ColorReset)
			} else {
				fmt.Printf("%s[Notification] Backup sent to Telegram successfully!%s\n", constant.ColorGreen, constant.ColorReset)
			}
		}

		fmt.Printf("%s[System] Backup completed successfully!%s\n", constant.ColorGreen, constant.ColorReset)
		return nil
	}

	if cfg.Schedule != "" {
		c := cron.New()
		var entryID cron.EntryID
		var err error

		entryID, err = c.AddFunc(cfg.Schedule, func() {
			now := time.Now()
			fmt.Printf("\n%s[Cron] [Triggered] Execution Time: %s%s\n",
				constant.ColorPurple, now.Format("2006-01-02 15:04:05"), constant.ColorReset)

			currentCfg := cfg
			currentCfg.Output = generateTimestampedOutput(cfg.Output, now)

			fmt.Printf("%s[Cron] Starting scheduled backup: %s%s\n",
				constant.ColorBlue, currentCfg.Output, constant.ColorReset)

			if err := processAction(currentCfg); err != nil {
				if ctx.Err() == context.Canceled {
					fmt.Printf("%s[Cron] Backup aborted by user.%s\n", constant.ColorRed, constant.ColorReset)
				} else {
					fmt.Printf("%s[Cron] Backup failed: %v%s\n", constant.ColorRed, err, constant.ColorReset)
				}
			} else {
				nextRun := c.Entry(entryID).Next
				fmt.Printf("%s[Cron] Task finished. Next run scheduled at: %s%s\n",
					constant.ColorGray, nextRun.Format("2006-01-02 15:04:05"), constant.ColorReset)
			}
		})

		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}

		c.Start()
		fmt.Printf("%s[Cron] System Active! Schedule: %s%s\n", constant.ColorGreen, cfg.Schedule, constant.ColorReset)
		fmt.Printf("%s[Cron] First scheduled run will be at: %s%s\n", constant.ColorCyan, c.Entry(entryID).Next.Format("2006-01-02 15:04:05"), constant.ColorReset)
		fmt.Println("[Cron] Press Ctrl+C to stop the application.")

		<-ctx.Done()
		fmt.Printf("%s[Cron] Stopping scheduler...%s\n", constant.ColorYellow, constant.ColorReset)

		stopCtx := c.Stop()
		<-stopCtx.Done()

		fmt.Printf("%s[Cron] Scheduler stopped. Exiting application.%s\n", constant.ColorGreen, constant.ColorReset)
		fmt.Printf("%s[System] Shutdown complete. Goodbye!%s\n", constant.ColorRed, constant.ColorReset)

		return nil
	}

	fmt.Printf("%s[System] Running one-time backup...%s\n", constant.ColorBlue, constant.ColorReset)
	return processAction(cfg)
}

func executeBackup(ctx context.Context, cfg Config) error {
	var driver database.Database
	switch cfg.DBType {
	case "mysql":
		driver = &database.MySQL{}
	case "postgres":
		driver = &database.Postgres{}
	default:
		return fmt.Errorf("unsupported db")
	}

	conn, err := driver.Connect(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName)
	if err != nil {
		return err
	}
	defer conn.Close()

	if len(cfg.Tables) == 0 {
		tables, err := driver.ListTables(ctx, conn)
		if err != nil {
			return err
		}

		cfg.Tables = tables
	}

	zipWriter, err := zipper.NewZipWriter(cfg.Output)
	if err != nil {
		return err
	}

	defer func() {
		zipWriter.Close()
		if ctx.Err() != nil {
			os.Remove(cfg.Output)
		}
	}()

	progress := make(map[string]*TableProgress)
	var mu sync.Mutex
	var zipMu sync.Mutex
	doneUI := make(chan struct{})

	for _, t := range cfg.Tables {
		progress[t] = &TableProgress{Table: t}
	}

	go startRenderer(ctx, progress, &mu, len(cfg.Tables), doneUI)

	err = runWorkerPool(ctx, defaultWorkerCount, cfg.Tables, func(table string) error {
		mu.Lock()
		progress[table].StartTime = time.Now()
		mu.Unlock()

		pr, pw := io.Pipe()

		go func() {
			err := driver.DumpTable(ctx, conn, table, pw)
			pw.CloseWithError(err)
		}()

		reader := io.TeeReader(pr, writerFunc(func(n int) {
			mu.Lock()
			progress[table].Bytes += int64(n)
			mu.Unlock()
		}))

		if ctx.Err() != nil {
			return ctx.Err()
		}

		zipMu.Lock()
		addErr := zipWriter.AddFile(table+".sql", reader)
		zipMu.Unlock()

		if addErr != nil {
			return addErr
		}

		mu.Lock()
		progress[table].Done = true
		mu.Unlock()

		return nil
	})

	<-doneUI // Wait for UI to finish rendering the last frame

	if err != nil {
		return err
	}

	return zipWriter.Close()
}

type writerFunc func(int)

func (f writerFunc) Write(p []byte) (int, error) {
	f(len(p))
	return len(p), nil
}

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
