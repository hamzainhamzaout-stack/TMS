package domain

import (
	"fmt"
	"strings"
	"time"
)

const customLayout = "2006-01-02T15:04"

type MyTime time.Time

func (t *MyTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}
	parsed, err := time.Parse(customLayout, s)
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
	return []byte(fmt.Sprintf(`"%s"`, time.Time(t).Format(customLayout))), nil
}

// String implements the Stringer interface for easy logging/printing
func (t MyTime) String() string {
	return time.Time(t).Format(customLayout)
}
