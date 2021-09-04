package keyword

import (
	"context"
	"fmt"
	"strings"
)

type Counter struct{}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Count(_ context.Context, text string) (map[string]int, error) {
	text, err := clean(text)
	if err != nil {
		return nil, fmt.Errorf("failed cleaning up the text: %w", err)
	}

	text = strings.ToLower(text)

	keywords := strings.Split(text, " ")
	out := make(map[string]int)

	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}

		out[keyword]++
	}

	if len(out) == 0 {
		out = nil
	}

	return out, err
}

func clean(text string) (string, error) {
	stopChars := ".?![](){}+=\"';,><@#$%^&*~|\\/:1234567890"

	sb := &strings.Builder{}
	sb.Grow(len(text))

	for _, c := range text {
		if !strings.ContainsRune(stopChars, c) {
			if _, err := sb.WriteRune(c); err != nil {
				return "", err
			}
		}
	}

	return strings.TrimSpace(sb.String()), nil
}
