package seed

import (
	"fmt"
	"time"

	"employee-profile-api/models"
	"gorm.io/gorm"
)

// Employees adds 25 fictional profiles only when the employee table is empty.
func Employees(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Employee{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	type person struct{ name, lastname, birthday string }
	people := []person{
		{"Narin", "Sample", "2004-02-12"}, {"Mali", "Demo", "2006-07-23"}, {"ธนา", "ตัวอย่าง", "2008-01-05"}, {"พิมพ์", "ทดสอบ", "2010-11-14"}, {"Arun", "Practice", "2014-04-27"},
		{"Kamon", "Sample", "1987-03-16"}, {"Lina", "Demo", "1990-09-08"}, {"สมชาย", "ตัวอย่าง", "1993-12-20"}, {"นิดา", "ทดสอบ", "1998-06-30"}, {"Mason", "Practice", "2001-10-09"},
		{"Anan", "Sample", "1971-05-19"}, {"Nora", "Demo", "1975-08-02"}, {"กานต์", "ตัวอย่าง", "1978-01-25"}, {"วิภา", "ทดสอบ", "1981-11-03"}, {"Ravi", "Practice", "1984-04-18"},
		{"Prasert", "Sample", "1953-02-06"}, {"Helen", "Demo", "1957-09-13"}, {"มานี", "ตัวอย่าง", "1961-12-22"}, {"ปรีชา", "ทดสอบ", "1965-06-04"}, {"Grace", "Practice", "1969-10-17"},
		{"Somsak", "Sample", "1928-03-01"}, {"Edith", "Demo", "1932-07-12"}, {"บุญมี", "ตัวอย่าง", "1938-01-29"}, {"ละไม", "ทดสอบ", "1943-11-21"}, {"Walter", "Practice", "1948-05-15"},
	}
	now := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		for i, p := range people {
			e := models.Employee{Name: p.name, Lastname: p.lastname, Birthday: p.birthday,
				Email: seedEmail(i), Tel: seedTel(i)}
			e.Age = e.CurrentAge(now)
			if err := tx.Create(&e).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func seedEmail(i int) string {
	return fmt.Sprintf("employee%02d@example.com", i+1)
}

func seedTel(i int) string {
	return fmt.Sprintf("080000%04d", i+1)
}
