package cache

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const cacheDir = ".cache"

func cachePath(key string) string {
	return path.Join(cacheDir, key+".json.gz")
}

func Exists(key string) bool {
	_, err := os.Stat(cachePath(key))
	return err == nil
}

func Load[T any](key string) (T, error) {
	var result T

	fpath := cachePath(key)
	f, err := os.Open(fpath)
	if err != nil {
		return result, fmt.Errorf("unable to open cache file: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return result, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	decoder := json.NewDecoder(gz)
	err = decoder.Decode(&result)
	if err != nil {
		return result, fmt.Errorf("json decode: %w", err)
	}

	return result, nil
}

func Save(key string, data any) error {
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}

	fpath := cachePath(key)
	f, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	encoder := json.NewEncoder(gz)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}
