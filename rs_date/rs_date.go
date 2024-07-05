package rs_date

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type RSDate struct {
	time.Time
}

func (d RSDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

func (d *RSDate) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`) // Remove quotes
	newTime, err := time.Parse("2006-01-02", strInput)
	if err != nil {
		return err
	}
	d.Time = newTime
	return nil
}

func (d *RSDate) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case []byte:
		var err error
		d.Time, err = time.Parse("2006-01-02", string(v))
		if err != nil {
			return err
		}
		return nil
	case string:
		var err error
		d.Time, err = time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("not a valid date")
	}
}

func (d *RSDate) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format("2006-01-02"), nil
}
