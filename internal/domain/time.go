package domain

import (
	"strings"
	"time"
)

type MyTime time.Time

func (t *MyTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	parsed, err := time.Parse("2006-01-02T15:04", s)
	if err != nil {
		return err
	}
	*t = MyTime(parsed)
	return nil
}
