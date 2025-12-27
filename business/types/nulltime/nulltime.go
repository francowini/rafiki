// Package nulltime provides a nullable time.Time type for business domain models.
package nulltime

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Null represents a nullable time.Time for optional timestamp fields.
// It implements sql.Scanner and driver.Valuer for database compatibility.
type Null struct {
	Time  time.Time
	Valid bool // Valid is true if Time is set
}

// New creates a valid Null with the given time.
func New(t time.Time) Null {
	return Null{
		Time:  t,
		Valid: true,
	}
}

// FromSQLNullTime converts sql.NullTime to Null.
func FromSQLNullTime(nt sql.NullTime) Null {
	return Null{
		Time:  nt.Time,
		Valid: nt.Valid,
	}
}

// ToSQLNullTime converts Null to sql.NullTime.
func (n Null) ToSQLNullTime() sql.NullTime {
	return sql.NullTime{
		Time:  n.Time,
		Valid: n.Valid,
	}
}

// Value returns the time.Time value, or zero time if null.
func (n Null) Value() time.Time {
	if !n.Valid {
		return time.Time{}
	}
	return n.Time
}

// String returns the RFC3339 string representation, or empty string if null.
func (n Null) String() string {
	if !n.Valid {
		return ""
	}
	return n.Time.Format(time.RFC3339)
}

// IsNull returns true if this Null represents a null value.
func (n Null) IsNull() bool {
	return !n.Valid
}

// Equal provides support for the go-cmp package and testing.
func (n Null) Equal(n2 Null) bool {
	if n.Valid != n2.Valid {
		return false
	}
	if !n.Valid {
		return true // Both are null
	}
	return n.Time.Equal(n2.Time)
}

// MarshalText provides support for logging and any marshal needs.
func (n Null) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return []byte(n.Time.Format(time.RFC3339)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// Treats "null" or empty string as a null value.
func (n *Null) UnmarshalText(data []byte) error {
	s := string(data)

	// Treat "null" or empty string as null
	if s == "" || s == "null" {
		*n = Null{}
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*n = New(t)
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (n Null) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Time)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *Null) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = Null{}
		return nil
	}

	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*n = New(t)
	return nil
}

// Scan implements the sql.Scanner interface for database reads.
func (n *Null) Scan(value any) error {
	if value == nil {
		*n = Null{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*n = New(v)
		return nil
	default:
		// Let sql.NullTime handle the conversion
		var nt sql.NullTime
		if err := nt.Scan(value); err != nil {
			return err
		}
		*n = FromSQLNullTime(nt)
		return nil
	}
}

// DriverValue implements the driver.Valuer interface for database writes.
func (n Null) DriverValue() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}
