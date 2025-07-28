package rs_date

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RSDate struct {
	time.Time
}

func (d RSDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

func NewRSDate(year int, month time.Month, day int) RSDate {
	return RSDate{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

var dateFormats = []string{
	"2006-01-02",
	"02/01/2006",
	"02/01/06",
	"2006/01/02",
	"02-Jan-06",
	"01/02/2006",
	"01/02/06",
	"02-Jan-2006",
	"02 Jan 2006",
	"02 January 2006",
	"January 02, 2006",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05.999Z07:00",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"02 Jan 06 15:04 MST",
	"Mon Jan _2 15:04:05 2006",
	"Mon Jan _2 15:04:05 MST 2006",
	"Mon Jan 02 15:04:05 -0700 2006",
	"2006-01-02 15:04:05 -0700 MST",
	"Jan-06",
}

func (d *RSDate) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`) // Remove quotes
	/*newTime, err := time.Parse("2006-01-02", strInput)
	if err != nil {
		return err
	}
	d.Time = newTime
	return nil*/
	for _, format := range dateFormats {
		newTime, err := time.Parse(format, strInput)
		if err == nil {
			d.Time = newTime
			return nil
		}
	}
	return errors.New("invalid date format")
}

/*
Early Unix timestamps (0–60,000) will be treated as Excel serial dates.
This is usually fine, since in practice, Unix timestamps in that range are rare
(they represent only the first day or two of 1970), but if you expect to handle such early Unix timestamps,
you may need to add additional logic or a format hint to distinguish them.
*/
func (d *RSDate) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	var str string
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case []byte:
		str = string(v)
	case string:
		str = v
	case float64:
		// Excel serial date as float64
		if v > 0 && v < 60000 {
			excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			d.Time = excelEpoch.Add(time.Duration(v*24) * time.Hour)
			return nil
		}
		return fmt.Errorf("float64 value out of Excel date range: %v", v)
	case int, int32, int64:
		var n int64
		switch t := v.(type) {
		case int:
			n = int64(t)
		case int32:
			n = int64(t)
		case int64:
			n = t
		}
		// Try as Excel serial date
		if n > 0 && n < 60000 {
			excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			d.Time = excelEpoch.Add(time.Duration(float64(n)*24) * time.Hour)
			return nil
		}
		// Try as Unix timestamp
		if n > 1000000000 && n < 2000000000 {
			d.Time = time.Unix(n, 0).UTC()
			return nil
		}
		return fmt.Errorf("int value not recognized as date: %v", n)
	default:
		return fmt.Errorf("not a valid date")
	}
	str = strings.TrimSpace(str)
	for _, format := range dateFormats {
		t, err := time.Parse(format, str)
		if err == nil {
			d.Time = t
			return nil
		}
	}
	if excelFloat, err := strconv.ParseFloat(str, 64); err == nil {
		if excelFloat > 0 && excelFloat < 60000 {
			excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			d.Time = excelEpoch.Add(time.Duration(excelFloat*24) * time.Hour)
			return nil
		}
	}
	if unixInt, err := strconv.ParseInt(str, 10, 64); err == nil {
		if unixInt > 1000000000 && unixInt < 2000000000 {
			d.Time = time.Unix(unixInt, 0).UTC()
			return nil
		}
	}
	return fmt.Errorf("not a valid date: %v", value)
}

func (d RSDate) Value() (driver.Value, error) {
	//if d == nil {
	//	return nil, nil
	//}
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format("2006-01-02"), nil
}

func (d *RSDate) GetLastDayOfMonth() RSDate {
	year, month, _ := d.Date()
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, d.Location())
	lastOfMonth := firstOfNextMonth.AddDate(0, 0, -1)
	return RSDate{Time: lastOfMonth}
}
