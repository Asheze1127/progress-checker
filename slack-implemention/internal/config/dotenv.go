package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadDotEnv(paths ...string) error {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if err := loadSingleDotEnv(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func loadSingleDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid .env format at %s:%d", path, lineNum)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("empty key at %s:%d", path, lineNum)
		}

		// Keep existing environment variables prioritized over .env values.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = trimQuotes(value)
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set env %q: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan %s: %w", path, err)
	}

	return nil
}

func trimQuotes(value string) string {
	if len(value) < 2 {
		return value
	}

	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}

	return value
}
