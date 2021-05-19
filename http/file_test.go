package http

import (
	"fmt"
	"regexp"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
)

func TestGenerateFilename(t *testing.T) {
	cases := []struct {
		count  int
		name   string
		result string
	}{
		{0, "foobar", `^foobar$`},
		{1, "foobar", `^foobar \(1\)$`},
		{5, "foobar", `^foobar \(5\)$`},
		{1, "foobar.zip", `^foobar \(1\).zip$`},
		{3, "foobar.zip", `^foobar \(3\).zip$`},
		{3, "funny #4 1234.kkk213.zip", `^funny #4 1234\.kkk213 \(3\)\.zip$`},
		{20, "foobar.zip", `^foobar_\w+.zip$`},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%d_%s", c.count, c.name), func(t *testing.T) {
			atts := make(map[string]models.Attachment, c.count)
			for i := 0; i < c.count; i++ {
				name := c.name
				if i > 0 {
					name = numberSuffix(name, i)
				}
				atts[name] = models.Attachment{Name: name}
			}
			result := generateFilename(c.name, atts)
			if m, err := regexp.MatchString(c.result, result); !m || err != nil {
				t.Errorf("makeUniqueFilename(%q, %v) = %q; expected to match %v",
					c.name, atts, result, c.result)
			}
		})
	}

}
