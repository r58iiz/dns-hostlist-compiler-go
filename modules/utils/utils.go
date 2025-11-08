package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isLocalFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func Download(fileURL string) (string, error) {
	var fileContent string

	if isURL(fileURL) {
		client := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}

		resp, err := client.Get(fileURL)
		if err != nil {
			return "", fmt.Errorf("error while fetching %s:\n%w", fileURL, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error while reading %s:\n%w", fileURL, err)
		}
		fileContent = string(body)
	} else if isLocalFile(fileURL) {
		file, err := os.Open(fileURL)
		if err != nil {
			return "", fmt.Errorf("error while opening local file %s: %w", fileURL, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}

		fileContent = string(content)
	} else {
		return "", fmt.Errorf("invalid URL or file path: %s", fileURL)
	}

	return fileContent, nil
}

func SubstringBetween(str string, startTag string, endTag string) string {
	if len(str) == 0 {
		return ""
	}

	var start int = strings.Index(str, startTag) + len(startTag)
	var end int = strings.Index(str, endTag)
	if end > start && start != -1 {
		return str[start:end]
	}

	return ""
}

func SplitByDelimiterWithEscapeCharacter(str string, delimiter byte, escapeCharacter byte, preserveAllTokens bool) []string {
	var parts []string

	if len(str) == 0 {
		return parts
	}

	var sb []byte
	for i := 0; i < len(str); i += 1 {
		var c byte = str[i]
		if c == delimiter {
			if i == 0 {
				// Ignore
			} else if str[i-1] == escapeCharacter {
				sb = sb[:len(sb)-1]
				sb = append(sb, c)
			} else if preserveAllTokens || len(sb) > 0 {
				parts = append(parts, string(sb))
				sb = []byte{}
			}
		} else {
			sb = append(sb, c)
		}
	}

	if preserveAllTokens || len(sb) > 0 {
		parts = append(parts, string(sb))
	}

	return parts
}

type Wildcard struct {
	regex    *regexp.Regexp
	plainStr string
}

func NewWildcard(str string) (*Wildcard, error) {
	if str == "" {
		return nil, fmt.Errorf("Wildcard/New - wildcard cannot be empty %s", str)
	}

	w := &Wildcard{
		regex:    nil,
		plainStr: str,
	}

	if strings.HasPrefix(str, "/") && strings.HasSuffix(str, "/") && len(str) > 2 {
		var re string = str[1 : len(str)-1]
		regex, err := regexp.Compile(re)
		if err != nil {
			return nil, err
		}
		w.regex = regex
	} else if strings.Contains(str, "*") {
		escapedStr := regexp.QuoteMeta(str)
		escapedStr = strings.ReplaceAll(escapedStr, "\\*", ".*")
		escapedStr = fmt.Sprintf("^%s$", escapedStr)
		regex, err := regexp.Compile(escapedStr)
		if err != nil {
			return nil, err
		}
		w.regex = regex
	}

	return w, nil
}

func (w Wildcard) Test(str string) bool {
	if w.regex != nil {
		return w.regex.Match([]byte(str))
	}

	return strings.Contains(str, w.plainStr)
}

func (w Wildcard) ToString() string {
	return w.plainStr
}
