package lib

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

// TODO: Learn strings.Builder

// FormatFileName generates a formatted file name based on the given tweet and format string.
func FormatFileName(tweet *twitterscraper.Tweet, format string, dateFormat string) (string, error) {
	formatParts := strings.Split(format, " ")
	formattedName := make([]string, 0, len(formatParts))

	invalidCharsRegex, err := regexp.Compile(`[/\\:*?"<>|]`)
	if err != nil {
		return "", fmt.Errorf("error compiling regular expression: %w", err)
	}

	for _, part := range formatParts {
		formatted, err := formatPart(tweet, part, invalidCharsRegex, dateFormat)
		if err != nil {
			return "", err
		}
		formattedName = append(formattedName, formatted)
	}

	return strings.Join(formattedName, "_"), nil
}

func formatPart(tweet *twitterscraper.Tweet, part string, invalidCharsRegex *regexp.Regexp, dateFormat string) (string, error) {
	switch part {
	case "{DATE}":
		return formatDate(tweet.Timestamp, dateFormat)
	case "{NAME}":
		return tweet.Name, nil
	case "{USERNAME}":
		return tweet.Username, nil
	case "{TITLE}":
		return formatTitle(tweet, invalidCharsRegex)
	case "{ID}":
		return tweet.ID, nil
	default:
		return "", fmt.Errorf("invalid format part: %s", part)
	}
}

func formatDate(timestamp int64, dateFormat string) (string, error) {
	t := time.Unix(timestamp, 0)
	return t.Format(dateFormat), nil
}

func formatTitle(tweet *twitterscraper.Tweet, invalidCharsRegex *regexp.Regexp) (string, error) {
	text := strings.ReplaceAll(tweet.Text, "/", "_")
	remainingChars := 251 - len(tweet.ID) - 4

	if text == "" {
		return "", nil
	}

	if len(text) <= remainingChars {
		return text, nil
	}

	return processText(text, remainingChars, invalidCharsRegex), nil
}

func processText(text string, remainingChars int, invalidCharsRegex *regexp.Regexp) string {
	var result strings.Builder

	for _, char := range text {
		if remainingChars <= 0 {
			break
		}

		charStr := string(char)
		if invalidCharsRegex.MatchString(charStr) {
			if remainingChars > 1 {
				result.WriteRune('_')
				remainingChars--
			}
		} else {
			result.WriteRune(char)
			remainingChars--
		}
	}

	return result.String()
}
