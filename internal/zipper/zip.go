package zipper

import (
	"archive/zip"
	"io"
	"os"
	"sync"
)

type ZipWriter struct {
	writer *zip.Writer
	file   *os.File
	mu     sync.Mutex
}

func NewZipWriter(output string) (*ZipWriter, error) {
	f, err := os.Create(output)
	if err != nil {
		return nil, err
	}

	return &ZipWriter{
		writer: zip.NewWriter(f),
		file:   f,
	}, nil
}

func (z *ZipWriter) AddFile(name string, reader io.Reader) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	w, err := z.writer.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, reader)
	return err
}

func (z *ZipWriter) Close() error {
	if err := z.writer.Close(); err != nil {
		z.file.Close()
		return err
	}

	return z.file.Close()
}
