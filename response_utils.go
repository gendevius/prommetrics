package prommetrics

import (
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
)

func isSuccessStatus(status int) bool {
	return status >= 200 && status < 400
}

func extractEndpoint(path string) string {
	cleanPath := strings.Trim(path, "/")
	if cleanPath == "" {
		return "/"
	}

	parts := strings.Split(cleanPath, "/")

	newParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if isUuid(part) {
			newParts = append(newParts, "<uuid>")
			continue
		}
		if isInt(part) {
			newParts = append(newParts, "<int>")
			continue

		}

		newParts = append(newParts, part)
	}

	url := strings.Join(newParts, "/")

	return url
}

func isUuid(s string) bool {
	_, err := uuid.FromString(s)
	return err == nil
}

func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
