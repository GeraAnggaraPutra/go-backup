package notification

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func SendToTelegram(token, chatID, filePath, message string) error {
	method := "sendDocument"
	if filePath == "" {
		method = "sendMessage"
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("chat_id", chatID)
	_ = writer.WriteField("parse_mode", "HTML")

	if filePath != "" {
		_ = writer.WriteField("caption", message)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile("document", filepath.Base(filePath))
		if err != nil {
			return err
		}
		_, _ = io.Copy(part, file)
	} else {
		_ = writer.WriteField("text", message)
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}
