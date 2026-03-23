package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/infrastructure/githubclient"
)

func main() {
	ctx := context.Background()
	loadDotEnvFiles(".env", "backend/.env")

	owner := requiredEnv("GITHUB_OWNER")
	repo := requiredEnv("GITHUB_REPO")
	title := requiredEnv("GITHUB_ISSUE_TITLE")

	client, err := githubclient.NewClient(loadGitHubConfig())
	if err != nil {
		log.Fatal(err)
	}

	createdIssue, err := client.CreateIssue(ctx, githubclient.CreateIssueInput{
		Owner:  owner,
		Repo:   repo,
		Title:  title,
		Body:   os.Getenv("GITHUB_ISSUE_BODY"),
		Labels: splitCSV(os.Getenv("GITHUB_ISSUE_LABELS")),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("created issue #%d: %s", createdIssue.Number, createdIssue.URL)
}

func loadDotEnvFiles(paths ...string) {
	for _, path := range paths {
		if err := loadDotEnvFile(path); err != nil {
			log.Fatal(err)
		}
	}
}

func loadDotEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")

	for index := 0; index < len(lines); {
		key, value, ok, nextIndex, parseErr := parseDotEnvEntry(lines, index)
		if parseErr != nil {
			return fmt.Errorf("parse %s line %d: %w", path, index+1, parseErr)
		}

		index = nextIndex

		if !ok {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s from %s: %w", key, path, err)
		}
	}

	return nil
}

func parseDotEnvEntry(lines []string, startIndex int) (string, string, bool, int, error) {
	rawLine := lines[startIndex]
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, startIndex + 1, nil
	}

	line = strings.TrimPrefix(line, "export ")

	separatorIndex := strings.IndexRune(line, '=')
	if separatorIndex < 0 {
		return "", "", false, startIndex + 1, fmt.Errorf("missing '=' separator")
	}

	key := strings.TrimSpace(line[:separatorIndex])
	if key == "" {
		return "", "", false, startIndex + 1, fmt.Errorf("env key is empty")
	}

	value := strings.TrimSpace(line[separatorIndex+1:])
	if value == "" {
		return key, value, true, startIndex + 1, nil
	}

	if isPEMBlockStart(value) {
		valueLines := []string{value}

		for index := startIndex + 1; index < len(lines); index++ {
			currentLine := strings.TrimRight(lines[index], "\r")
			valueLines = append(valueLines, currentLine)
			if isPEMBlockEnd(strings.TrimSpace(currentLine)) {
				return key, strings.Join(valueLines, "\n"), true, index + 1, nil
			}
		}

		return "", "", false, startIndex + 1, fmt.Errorf("unterminated pem value")
	}

	if value[0] != '"' && value[0] != '\'' {
		return key, value, true, startIndex + 1, nil
	}

	quote := value[0]
	if len(value) >= 2 && value[len(value)-1] == quote {
		return key, value[1 : len(value)-1], true, startIndex + 1, nil
	}

	valueLines := []string{value[1:]}

	for index := startIndex + 1; index < len(lines); index++ {
		currentLine := lines[index]
		if strings.HasSuffix(currentLine, string(quote)) {
			valueLines = append(valueLines, currentLine[:len(currentLine)-1])
			return key, strings.Join(valueLines, "\n"), true, index + 1, nil
		}

		valueLines = append(valueLines, currentLine)
	}

	return "", "", false, startIndex + 1, fmt.Errorf("unterminated quoted value")
}

func isPEMBlockStart(value string) bool {
	return strings.HasPrefix(value, "-----BEGIN ") && strings.HasSuffix(value, "-----")
}

func isPEMBlockEnd(value string) bool {
	return strings.HasPrefix(value, "-----END ") && strings.HasSuffix(value, "-----")
}

func loadGitHubConfig() githubclient.Config {
	appIssuer := firstNonEmptyEnv("GITHUB_APP_CLIENT_ID", "GITHUB_APP_ID")
	appPrivateKey := loadAppPrivateKey()
	appInstallationID, err := optionalInt64Env("GITHUB_APP_INSTALLATION_ID")
	if err != nil {
		log.Fatal(err)
	}

	if appIssuer != "" || appPrivateKey != "" || appInstallationID != 0 {
		if appIssuer == "" {
			log.Fatal("GITHUB_APP_CLIENT_ID or GITHUB_APP_ID is required for GitHub App authentication")
		}

		if appPrivateKey == "" {
			log.Fatal("GITHUB_APP_PRIVATE_KEY_PEM is required for GitHub App authentication")
		}

		return githubclient.Config{
			BaseURL:           strings.TrimSpace(os.Getenv("GITHUB_BASE_URL")),
			AppIssuer:         appIssuer,
			AppInstallationID: appInstallationID,
			AppPrivateKeyPEM:  appPrivateKey,
		}
	}

	return githubclient.Config{
		BaseURL: strings.TrimSpace(os.Getenv("GITHUB_BASE_URL")),
		Token:   requiredEnv("GITHUB_TOKEN"),
	}
}

func requiredEnv(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		log.Fatalf("%s is required", name)
	}

	return value
}

func firstNonEmptyEnv(names ...string) string {
	for _, name := range names {
		value := strings.TrimSpace(os.Getenv(name))
		if value != "" {
			return value
		}
	}

	return ""
}

func loadAppPrivateKey() string {
	return strings.TrimSpace(os.Getenv("GITHUB_APP_PRIVATE_KEY_PEM"))
}

func optionalInt64Env(name string) (int64, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return 0, nil
	}

	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return parsedValue, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	rawValues := strings.Split(value, ",")
	values := make([]string, 0, len(rawValues))

	for _, rawValue := range rawValues {
		trimmedValue := strings.TrimSpace(rawValue)
		if trimmedValue == "" {
			continue
		}

		values = append(values, trimmedValue)
	}

	if len(values) == 0 {
		return nil
	}

	return values
}
