package filesservice

import (
	"os"
	"strconv"
)

type Config struct {
	MaxSize          int64
	UploadsDir       string
	AllowedMimeTypes []string
}

func ConfigDefault() *Config {
	dir := os.Getenv("APP_UPLOADS_DIR")
	if dir == "" {
		dir = "./assets/uploads"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic(err)
		}
	}

	maxSize, err := strconv.ParseInt(os.Getenv("APP_UPLOADS_MAX_SIZE"), 10, 64)
	if err != nil || maxSize <= 0 {
		maxSize = 10 << 20
	}

	return &Config{
		MaxSize:    maxSize,
		UploadsDir: dir,
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
		},
	}
}
