package scalar // Changed from package graph

import (
	"fmt"
	"io"
	"strconv"
	"time"
	"regexp"
)

// Date custom scalar type
type Date struct {
	time.Time
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (d *Date) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("Date must be a string")
	}
	// Assuming date format YYYY-MM-DD
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return fmt.Errorf("Date must be in YYYY-MM-DD format: %w", err)
	}
	d.Time = t
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (d Date) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, "%q", d.Time.Format("2006-01-02"))
}

// Email custom scalar type
type Email string

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (e *Email) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("Email must be a string")
	}
	// Basic email validation regex
	// More comprehensive validation should be used in a real application
	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(str) {
		return fmt.Errorf("%s is not a valid Email", str)
	}
	*e = Email(str)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (e Email) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, "%s", strconv.Quote(string(e)))
}
