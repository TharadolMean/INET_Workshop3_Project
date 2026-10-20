package handlers

import (
	"errors"
	"strings"
	"time"

	"employee-profile-api/models"
	"employee-profile-api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeHandler struct{ DB *gorm.DB }

func (h EmployeeHandler) Create(c *fiber.Ctx) error {
	in, err := parseInput(c)
	if err != nil {
		return badJSON(c)
	}
	now := time.Now()
	_, details := utils.ValidateEmployee(in, now)
	if len(details) != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed", "details": details})
	}
	employee := models.Employee{Name: in.Name, Lastname: in.Lastname, Birthday: in.Birthday, Email: in.Email, Tel: in.Tel}
	employee.Age = employee.CurrentAge(now)
	if err := h.DB.Create(&employee).Error; err != nil {
		return databaseError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "employee created", "data": employee.Public(now)})
}

func (h EmployeeHandler) List(c *fiber.Ctx) error {
	var employees []models.Employee
	if err := h.DB.Order("id ASC").Find(&employees).Error; err != nil {
		return databaseError(c, err)
	}
	now := time.Now()
	data := make([]models.Response, 0, len(employees))
	for _, e := range employees {
		data = append(data, e.Public(now))
	}
	return c.JSON(fiber.Map{"data": data, "count": len(data)})
}

func (h EmployeeHandler) Get(c *fiber.Ctx) error {
	id, ok := validUUID(c.Params("id"))
	if !ok {
		return invalidUUID(c)
	}
	var employee models.Employee
	if err := h.DB.Where("employee_id = ?", id).First(&employee).Error; err != nil {
		return lookupError(c, err)
	}
	return c.JSON(fiber.Map{"data": employee.Public(time.Now())})
}

func (h EmployeeHandler) Update(c *fiber.Ctx) error {
	id, ok := validUUID(c.Params("id"))
	if !ok {
		return invalidUUID(c)
	}
	in, err := parseInput(c)
	if err != nil {
		return badJSON(c)
	}
	now := time.Now()
	_, details := utils.ValidateEmployee(in, now)
	if len(details) != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed", "details": details})
	}
	var employee models.Employee
	if err := h.DB.Where("employee_id = ?", id).First(&employee).Error; err != nil {
		return lookupError(c, err)
	}
	employee.Name, employee.Lastname, employee.Birthday = in.Name, in.Lastname, in.Birthday
	employee.Email, employee.Tel = in.Email, in.Tel
	employee.Age = employee.CurrentAge(now)
	if err := h.DB.Save(&employee).Error; err != nil {
		return databaseError(c, err)
	}
	return c.JSON(fiber.Map{"message": "employee updated", "data": employee.Public(now)})
}

func (h EmployeeHandler) Delete(c *fiber.Ctx) error {
	id, ok := validUUID(c.Params("id"))
	if !ok {
		return invalidUUID(c)
	}
	result := h.DB.Where("employee_id = ?", id).Delete(&models.Employee{})
	if result.Error != nil {
		return databaseError(c, result.Error)
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "employee not found"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h EmployeeHandler) Search(c *fiber.Ctx) error {
	keyword := strings.TrimSpace(c.Params("keyword"))
	if keyword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "keyword is required"})
	}
	// Escape LIKE metacharacters so the path keyword is matched literally.
	keyword = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(keyword)
	pattern := "%" + keyword + "%"
	var employees []models.Employee
	err := h.DB.Where(`employee_id LIKE ? ESCAPE '\' OR name LIKE ? ESCAPE '\' OR lastname LIKE ? ESCAPE '\'`, pattern, pattern, pattern).Order("id ASC").Find(&employees).Error
	if err != nil {
		return databaseError(c, err)
	}
	now := time.Now()
	data := make([]models.Response, 0, len(employees))
	for _, e := range employees {
		data = append(data, e.Public(now))
	}
	return c.JSON(fiber.Map{"data": data, "count": len(data)})
}

func Generation(age int) string {
	switch {
	case age < 24:
		return "gen_z"
	case age <= 41:
		return "gen_y"
	case age <= 56:
		return "gen_x"
	case age <= 75:
		return "baby_boomer"
	default:
		return "gi_generation"
	}
}

func (h EmployeeHandler) Generations(c *fiber.Ctx) error {
	var employees []models.Employee
	if err := h.DB.Order("id ASC").Find(&employees).Error; err != nil {
		return databaseError(c, err)
	}
	groups := map[string][]models.Response{
		"gen_z": {}, "gen_y": {}, "gen_x": {}, "baby_boomer": {}, "gi_generation": {},
	}
	now := time.Now()
	for _, e := range employees {
		p := e.Public(now)
		group := Generation(p.Age)
		groups[group] = append(groups[group], p)
	}
	return c.JSON(fiber.Map{"data": groups, "count": len(employees)})
}

func parseInput(c *fiber.Ctx) (utils.EmployeeInput, error) {
	var in utils.EmployeeInput
	err := c.BodyParser(&in)
	return in, err
}

func validUUID(raw string) (string, bool) {
	id, err := uuid.Parse(raw)
	return id.String(), err == nil && id.String() == raw
}

func invalidUUID(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid employee UUID"})
}

func badJSON(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
}

func lookupError(c *fiber.Ctx, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "employee not found"})
	}
	return databaseError(c, err)
}

func databaseError(c *fiber.Ctx, err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique constraint failed") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "employee already exists"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database operation failed"})
}
