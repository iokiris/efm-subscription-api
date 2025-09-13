package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Subscription общая структура для подписок

type Subscription struct {
	ID        int64      `db:"id" json:"id"`
	Service   string     `db:"service_name" json:"service_name"`
	Price     int        `db:"price" json:"price"`
	UserID    string     `db:"user_id" json:"user_id"`
	StartDate MonthYear  `db:"start_date" json:"start_date"`
	EndDate   *MonthYear `db:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// MonthYear — кастомный тип, необходимый для передачи MM-YYYY в валидный формат time.Time
type MonthYear time.Time

// UnmarshalJSON парсит данные из JSON в MonthYear.
//
// Вход: data = []byte(`"07-2025"`)
//
// Результат: MonthYear(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC))
func (my *MonthYear) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" {
		*my = MonthYear(time.Time{})
		return nil
	}
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return fmt.Errorf("invalid date format, expect MM-YYYY: %w", err)
	}
	*my = MonthYear(t)
	return nil
}

// MarshalJSON сериализует MonthYear в JSON строку формата "MM-YYYY"
//
// MonthYear(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)) -> []byte(`"07-2025"`)
func (my *MonthYear) MarshalJSON() ([]byte, error) {
	t := time.Time(*my)
	marshalled := []byte(fmt.Sprintf(`"%02d-%d"`, t.Month(), t.Year()))
	return marshalled, nil
}

func (my MonthYear) Value() (driver.Value, error) {
	t := time.Time(my)
	return t, nil // для совместимости с pgx
}

func (my *MonthYear) Scan(src interface{}) error {
	t, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan %T into MonthYear", src)
	}
	*my = MonthYear(t)
	return nil
}
