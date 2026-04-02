package selfupdate

import (
	"strconv"
	"strings"

	logger "github.com/sirupsen/logrus"
)

// CompareVersions compares two semantic versions.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
// Development builds ("dev") are always considered older than any release.
func CompareVersions(v1, v2 string) int {
	if v1 == "dev" {
		return -1
	}
	if v2 == "dev" {
		return 1
	}

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for _, part := range parts1 {
		if _, err := strconv.Atoi(part); err != nil {
			logger.Warnf(
				"Version %s contains non-numeric parts, cannot perform reliable comparison",
				v1,
			)
			return strings.Compare(v1, v2)
		}
	}
	for _, part := range parts2 {
		if _, err := strconv.Atoi(part); err != nil {
			logger.Warnf(
				"Version %s contains non-numeric parts, cannot perform reliable comparison",
				v2,
			)
			return strings.Compare(v1, v2)
		}
	}

	maxLen := max(len(parts2), len(parts1))

	for len(parts1) < maxLen {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < maxLen {
		parts2 = append(parts2, "0")
	}

	for i := range maxLen {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}
