package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"employee-profile-api/config"
	"employee-profile-api/handlers"
	"employee-profile-api/models"
	"employee-profile-api/routes"
	"employee-profile-api/seed"
	"github.com/google/uuid"
)

func TestAPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "employee.db")
	db, err := config.OpenDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	initialSQL, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = initialSQL.Close() })
	if err := seed.Employees(db); err != nil {
		t.Fatal(err)
	}
	if err := seed.Employees(db); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&models.Employee{}).Count(&count).Error; err != nil || count != 25 {
		t.Fatalf("seed count = %d, error = %v", count, err)
	}
	app := routes.NewApp(db, "testgo", "23012023")
	request := func(method, path, body string, auth bool) (int, map[string]any) {
		t.Helper()
		req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if auth {
			req.SetBasicAuth("testgo", "23012023")
		}
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNoContent {
			return resp.StatusCode, nil
		}
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, result
	}
	base := "http://localhost/api/v1/users"
	status, result := request(http.MethodGet, base, "", false)
	if status != 200 || result["count"] != float64(25) {
		t.Fatalf("public list: %d %v", status, result)
	}
	seenIDs := make(map[string]bool)
	for _, raw := range result["data"].([]any) {
		id := raw.(map[string]any)["employee_id"].(string)
		if _, err := uuid.Parse(id); err != nil || seenIDs[id] {
			t.Fatalf("invalid or duplicate seed UUID %q: %v", id, err)
		}
		seenIDs[id] = true
	}
	first := result["data"].([]any)[0].(map[string]any)
	if first["birthday"] != "2004-02-12" || first["age"] == nil {
		t.Fatalf("birthday or age missing: %v", first)
	}
	seedID := first["employee_id"].(string)
	if _, err := uuid.Parse(seedID); err != nil {
		t.Fatal(err)
	}
	status, _ = request(http.MethodGet, base+"/"+seedID, "", false)
	if status != 200 {
		t.Fatalf("public get = %d", status)
	}
	body := `{"name":"กานต์","lastname":"Example","birthday":"2000-05-15","email":"person@example.com","tel":"0812345678"}`
	status, _ = request(http.MethodPost, base, body, false)
	if status != 401 {
		t.Fatalf("unauthenticated post = %d", status)
	}
	status, _ = request(http.MethodPut, base+"/"+seedID, body, false)
	if status != 401 {
		t.Fatalf("unauthenticated put = %d", status)
	}
	status, _ = request(http.MethodDelete, base+"/"+seedID, "", false)
	if status != 401 {
		t.Fatalf("unauthenticated delete = %d", status)
	}
	for field, invalid := range map[string]string{
		"name":     `{"name":"John1","lastname":"Example","birthday":"2000-05-15","email":"person@example.com","tel":"0812345678"}`,
		"birthday": `{"name":"John","lastname":"Example","birthday":"2023-02-30","email":"person@example.com","tel":"0812345678"}`,
		"email":    `{"name":"John","lastname":"Example","birthday":"2000-05-15","email":"bad-email","tel":"0812345678"}`,
	} {
		status, response := request(http.MethodPost, base, invalid, true)
		if status != 400 || response["details"].(map[string]any)[field] == nil {
			t.Fatalf("invalid %s: %d %v", field, status, response)
		}
	}
	status, result = request(http.MethodPost, base, `{"name":"John","lastname":"Example","birthday":"2999-01-01","email":"person@example.com","tel":"0812345678"}`, true)
	if status != 400 || result["details"].(map[string]any)["birthday"] == nil {
		t.Fatalf("future birthday: %d %v", status, result)
	}
	status, result = request(http.MethodPost, base, body, true)
	if status != 201 {
		t.Fatalf("create: %d %v", status, result)
	}
	id := result["data"].(map[string]any)["employee_id"].(string)
	if _, err := uuid.Parse(id); err != nil {
		t.Fatal(err)
	}
	for _, keyword := range []string{id[:8], "กานต์", "Example"} {
		status, result = request(http.MethodGet, base+"/search/"+url.PathEscape(keyword), "", false)
		if status != 200 || result["count"].(float64) < 1 {
			t.Fatalf("search %q: %d %v", keyword, status, result)
		}
	}
	status, result = request(http.MethodGet, base+"/search/NoSuchEmployee", "", false)
	if status != 200 || result["count"] != float64(0) || len(result["data"].([]any)) != 0 {
		t.Fatalf("empty search: %d %v", status, result)
	}
	status, result = request(http.MethodGet, base+"/generations", "", false)
	if status != 200 || len(result["data"].(map[string]any)) != 5 {
		t.Fatalf("generations: %d %v", status, result)
	}
	for group, entries := range result["data"].(map[string]any) {
		if len(entries.([]any)) < 2 {
			t.Fatalf("group %s has fewer than 2 people", group)
		}
	}
	updated := strings.Replace(body, "กานต์", "มานะ", 1)
	status, result = request(http.MethodPut, base+"/"+id, updated, true)
	if status != 200 || result["data"].(map[string]any)["name"] != "มานะ" {
		t.Fatalf("update: %d %v", status, result)
	}
	status, result = request(http.MethodPut, base+"/"+id, `{"name":"Invalid1"}`, true)
	if status != 400 || result["details"].(map[string]any)["name"] == nil {
		t.Fatalf("invalid update: %d %v", status, result)
	}
	status, _ = request(http.MethodDelete, base+"/"+id, "", true)
	if status != 204 {
		t.Fatalf("delete = %d", status)
	}
	status, _ = request(http.MethodGet, base+"/"+id, "", false)
	if status != 404 {
		t.Fatalf("deleted get = %d", status)
	}
	status, _ = request(http.MethodGet, base+"/getuser", "", false)
	if status != 400 && status != 404 {
		t.Fatalf("unexpected verb route = %d", status)
	}
	if err := app.Shutdown(); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := config.OpenDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	reopenedSQL, err := reopened.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopenedSQL.Close() })
	if err := seed.Employees(reopened); err != nil {
		t.Fatal(err)
	}
	if err := reopened.Model(&models.Employee{}).Count(&count).Error; err != nil || count != 25 {
		t.Fatalf("restart count = %d, error = %v", count, err)
	}
}

func TestGenerationBoundaries(t *testing.T) {
	cases := map[int]string{23: "gen_z", 24: "gen_y", 41: "gen_y", 42: "gen_x", 56: "gen_x", 57: "baby_boomer", 75: "baby_boomer", 76: "gi_generation"}
	for age, want := range cases {
		t.Run(fmt.Sprint(age), func(t *testing.T) {
			if got := handlers.Generation(age); got != want {
				t.Fatalf("age %d: got %s, want %s", age, got, want)
			}
		})
	}
}
