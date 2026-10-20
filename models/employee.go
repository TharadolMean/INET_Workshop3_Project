package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Employee keeps GORM's internal primary key separate from the public UUID.
type Employee struct {
	gorm.Model `json:"-"`
	EmployeeID string `gorm:"type:text;uniqueIndex;not null" json:"employee_id"`
	Name       string `gorm:"not null" json:"name"`
	Lastname   string `gorm:"not null" json:"lastname"`
	Birthday   string `gorm:"type:text;not null" json:"birthday"`
	Age        int    `gorm:"not null" json:"age"`
	Email      string `gorm:"not null" json:"email"`
	Tel        string `gorm:"not null" json:"tel"`
}

func (e *Employee) BeforeCreate(_ *gorm.DB) error {
	if e.EmployeeID == "" {
		e.EmployeeID = uuid.NewString()
	}
	return nil
}

// CurrentAge computes age from the stored calendar date for API responses.
func (e Employee) CurrentAge(now time.Time) int {
	birth, err := time.Parse("2006-01-02", e.Birthday)
	if err != nil {
		return e.Age
	}
	age := now.Year() - birth.Year()
	if int(now.Month()) < int(birth.Month()) || (now.Month() == birth.Month() && now.Day() < birth.Day()) {
		age--
	}
	return age
}

type Response struct {
	EmployeeID string `json:"employee_id"`
	Name       string `json:"name"`
	Lastname   string `json:"lastname"`
	Birthday   string `json:"birthday"`
	Age        int    `json:"age"`
	Email      string `json:"email"`
	Tel        string `json:"tel"`
}

func (e Employee) Public(now time.Time) Response {
	return Response{e.EmployeeID, e.Name, e.Lastname, e.Birthday, e.CurrentAge(now), e.Email, e.Tel}
}
