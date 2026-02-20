package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
	"google.golang.org/api/option"
)

type GCS struct {
	Client     *storage.Client
	BucketName string
}

func NewStorage() (s *GCS, err error) {
	var (
		credentialFile = os.Getenv("GCS_SERVICE_ACCOUNT_FILE")
		bucketName     = os.Getenv("GCS_BUCKET_NAME")
	)

	client, err := storage.NewClient(context.Background(), option.WithCredentialsFile(credentialFile))
	if err != nil {
		return
	}

	s = &GCS{
		Client:     client,
		BucketName: bucketName,
	}

	return
}

func (g *GCS) UploadToGCS(ctx context.Context, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	wc := g.Client.Bucket(g.BucketName).Object(filePath).NewWriter(ctx)
	defer wc.Close()

	fmt.Printf("%s[Google Cloud Storage]%s Uploading %s%s%s to bucket %s%s%s...\n",
		constant.ColorCyan, constant.ColorReset,
		constant.ColorYellow, filePath, constant.ColorReset,
		constant.ColorGreen, g.BucketName, constant.ColorReset)

	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy failed: %v", err)
	}

	return nil
}
