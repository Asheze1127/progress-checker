package parser

import "strings"

func ParseKeyValueText(input string) map[string]string {
	result := make(map[string]string)
	for _, rawLine := range strings.Split(input, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		sep := strings.Index(line, ":")
		if sep <= 0 {
			continue
		}

		key := normalizeKey(line[:sep])
		value := strings.TrimSpace(line[sep+1:])
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}
	return result
}

func normalizeKey(key string) string {
	normalized := strings.ToLower(strings.TrimSpace(key))
	replacer := strings.NewReplacer(" ", "", "_", "", "-", "")
	return replacer.Replace(normalized)
}
