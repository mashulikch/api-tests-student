package helpers

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const onePixelPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jL5sAAAAASUVORK5CYII="

// CreatePhotoFile создает небольшой корректный PNG-файл, который служит тестовым материалом
// для проверки многокомпонентной (multipart) загрузки
func CreatePhotoFile(t *testing.T) string {
	t.Helper()

	image, err := base64.StdEncoding.DecodeString(onePixelPNG)
	require.NoError(t, err)
	photoPath := filepath.Join(t.TempDir(), "test-photo.png")
	require.NoError(t, os.WriteFile(photoPath, image, 0o600))
	return photoPath
}
