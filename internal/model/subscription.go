package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// YearMonth stores the first day of a calendar month and marshals as "MM-YYYY".
type YearMonth struct {
	time.Time
}

func NewYearMonth(year, month int) YearMonth {
	return YearMonth{Time: time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)}
}

func ParseYearMonth(s string) (YearMonth, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return YearMonth{}, fmt.Errorf("invalid date format, expected MM-YYYY: %w", err)
	}
	return YearMonth{Time: time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)}, nil
}

func (ym YearMonth) String() string {
	return ym.Format("01-2006")
}

func (ym YearMonth) MarshalJSON() ([]byte, error) {
	return json.Marshal(ym.String())
}

func (ym *YearMonth) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseYearMonth(s)
	if err != nil {
		return err
	}
	*ym = parsed
	return nil
}

// MonthsUntilInclusive counts months from ym to other inclusive.
func (ym YearMonth) MonthsUntilInclusive(other YearMonth) int {
	return (other.Year()-ym.Year())*12 + int(other.Month()-ym.Month()) + 1
}

func MaxYearMonth(a, b YearMonth) YearMonth {
	if a.Time.After(b.Time) {
		return a
	}
	return b
}

func MinYearMonth(a, b YearMonth) YearMonth {
	if a.Time.Before(b.Time) {
		return a
	}
	return b
}

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   YearMonth  `json:"start_date"`
	EndDate     *YearMonth `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   YearMonth  `json:"start_date"`
	EndDate     *YearMonth `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string    `json:"service_name,omitempty"`
	Price       *int       `json:"price,omitempty"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	StartDate   *YearMonth `json:"start_date,omitempty"`
	EndDate     *YearMonth `json:"end_date,omitempty"`
}

type ListFilter struct {
	UserID      uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

type ListResponse struct {
	Items  []Subscription `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type SumFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        YearMonth
	To          YearMonth
}

type SumResponse struct {
	Total       int        `json:"total"`
	From        YearMonth  `json:"from"`
	To          YearMonth  `json:"to"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	ServiceName *string    `json:"service_name,omitempty"`
}

// OverlapMonths returns how many months of the subscription fall into [from, to].
func OverlapMonths(start YearMonth, end *YearMonth, from, to YearMonth) int {
	subEnd := to
	if end != nil {
		subEnd = MinYearMonth(*end, to)
	}
	subStart := MaxYearMonth(start, from)
	if subStart.Time.After(subEnd.Time) {
		return 0
	}
	return subStart.MonthsUntilInclusive(subEnd)
}
