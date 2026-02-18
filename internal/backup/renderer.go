package backup

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
)

func startRenderer(ctx context.Context, progress map[string]*TableProgress, mu *sync.Mutex, total int, done chan struct{}) {
	defer close(done)
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	renderedLines := 0
	frameIdx := 0
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			mu.Lock()

			keys := make([]string, 0, len(progress))
			completedCount := 0
			for k, p := range progress {
				keys = append(keys, k)
				if p.Done {
					completedCount++
				}
			}

			// sort with unfinished tables first, then by name
			sort.Slice(keys, func(i, j int) bool {
				if progress[keys[i]].Done != progress[keys[j]].Done {
					return !progress[keys[i]].Done
				}
				return keys[i] < keys[j]
			})

			isFinished := completedCount == total

			// Clear previous frame
			if renderedLines > 0 {
				fmt.Printf(constant.CursorUp, renderedLines)
			}

			var frame strings.Builder
			sp := constant.ColorCyan + spinnerChars[frameIdx%len(spinnerChars)] + constant.ColorReset
			if isFinished {
				sp = constant.ColorGreen + constant.IconSuccess + constant.ColorReset
			}

			frame.WriteString(fmt.Sprintf("\r%s%s %sDATABASE BACKUP SYSTEM%s [%d/%d]\n",
				constant.ClearLine, sp, constant.ColorBold+constant.ColorCyan, constant.ColorReset, completedCount, total))

			currentLines := 1
			displayedRows := 0
			maxRows := 20
			if isFinished {
				maxRows = total
			}

			for _, k := range keys {
				p := progress[k]
				if displayedRows < maxRows {
					status := constant.ColorYellow + constant.IconRunning + " Running" + constant.ColorReset

					if p.Done {
						status = constant.ColorGreen + constant.IconSuccess + " Done   " + constant.ColorReset
					}

					bar := buildCyberBar(p.Done, p.Bytes, frameIdx)
					size := formatSize(p.Bytes)

					// calculating speed in MB/s
					speed := 0.0
					if p.Bytes > 0 {
						elapsed := time.Since(p.StartTime).Seconds()
						if elapsed < 0.001 {
							elapsed = 0.001
						}

						speed = float64(p.Bytes) / elapsed / 1024 / 1024
					}

					frame.WriteString(fmt.Sprintf("\r%s   %-30s %s   %-10s %s   %6.2f MB/s\n",
						constant.ClearLine,
						truncateString(p.Table, 30),
						bar,
						size,
						status,
						speed))

					currentLines++
					displayedRows++
				}
			}

			// If there are more tables than displayed, show a summary line
			if total > displayedRows && !isFinished {
				frame.WriteString(fmt.Sprintf("\r%s   %s... and %d other tables in queue ...%s\n",
					constant.ClearLine, constant.ColorGray, total-displayedRows, constant.ColorReset))
				currentLines++
			}

			// Progress summary line
			percent := (completedCount * 100) / total
			frame.WriteString(fmt.Sprintf("\r%s\n\r%s%sProgress: %d%% | Time: %s | Tables: %d/%d%s\n",
				constant.ClearLine, constant.ClearLine, constant.ColorBold+constant.ColorBlue,
				percent, time.Since(start).Round(time.Second), completedCount, total, constant.ColorReset))
			currentLines += 2

			fmt.Print(frame.String())
			renderedLines = currentLines
			frameIdx++

			if isFinished {
				fmt.Printf("\r%s\n%s%s✨ ALL BACKUPS COMPLETED SUCCESSFULLY IN %v ✨%s\n",
					constant.ClearLine, constant.ColorBold, constant.ColorGreen, time.Since(start).Round(time.Second), constant.ColorReset)
				mu.Unlock()
				return
			}
			mu.Unlock()
		}
	}
}

func truncateString(s string, l int) string {
	if len(s) > l {
		return s[:l-3] + "..."
	}

	return s
}

func buildCyberBar(done bool, bytes int64, frame int) string {
	const width = 25
	if done {
		return "[" + constant.ColorGreen + strings.Repeat("■", width) + constant.ColorReset + "]"
	}

	fillLen := int((bytes / (128 * 1024)) % int64(width))
	if fillLen == 0 && bytes > 0 {
		fillLen = 1
	}
	scanPos := frame % width

	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < width; i++ {
		if i < fillLen {
			if i == scanPos {
				bar.WriteString(constant.ColorBold + "\033[37m■" + constant.ColorReset)
			} else {
				bar.WriteString(constant.ColorCyan + "■" + constant.ColorReset)
			}
		} else {
			char := "░"
			if i == scanPos {
				char = "■"
			}
			bar.WriteString(constant.ColorGray + char + constant.ColorReset)
		}
	}

	bar.WriteString("]")
	return bar.String()
}

func formatSize(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	val := float64(bytes)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}

	return fmt.Sprintf("%.2f %s", val, units[i])
}
