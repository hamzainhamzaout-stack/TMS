package domain

import (
	"fmt"
	"strings"
	"time"
)

type MyTime time.Time

func (t *MyTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}
	layouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04"}
	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.Parse(layout, s)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	*t = MyTime(parsed)
	return nil
}

// func (t *MyTime) UnmarshalJSON(b []byte) error {
// 	s := strings.Trim(string(b), `"`)
// 	parsed, err := time.Parse("2006-01-02T15:04", s)
// 	if err != nil {
// 		return err
// 	}
// 	*t = MyTime(parsed)
// 	return nil
// }

func (t MyTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, time.Time(t).Format("2006-01-02T15:04"))), nil
}

// String implements the Stringer interface for easy logging/printing
func (t MyTime) String() string {
	return time.Time(t).Format("2006-01-02T15:04")
}
