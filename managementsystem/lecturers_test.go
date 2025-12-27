package managementsystem_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"managementsystem/managementsystem"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestHybridHandler5_CreateLecturersHandler(t *testing.T) {

	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MySQL_DSN", "root:root@tcp(127.0.0.1:3306)/management_sys")

	redisInstance, err := managementsystem.ConnectRedis()
	if err != nil {
		panic(err)
	}
	mysqlinstance, err := managementsystem.ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handler := &managementsystem.HybridHandler5{MySQL: mysqlinstance, Redis: redisInstance, Ctx: context.Background()}

	tests := []struct {
		name     string // description of this test case
		lecturer managementsystem.Lecturer
		willpass bool
	}{
		{
			name: "valide",
			lecturer: managementsystem.Lecturer{
				Name:        "Ramesh",
				Email:       "ramesh@gmail.com",
				Dept:        "CSE",
				Designation: "senior lecturer",
			},
			willpass: true,
		},
		{
			name: "invalid name and valid email  , dept, and designation",
			lecturer: managementsystem.Lecturer{
				Name:        "",
				Email:       "akash@gmail.com",
				Dept:        "CSE",
				Designation: "senior lecturer",
			},
			willpass: false,
		},
		{
			name: "invalid email and valid name , dept and designation",
			lecturer: managementsystem.Lecturer{
				Name:        "Akash",
				Email:       "",
				Dept:        "CSE",
				Designation: "senior lecturer",
			},
			willpass: false,
		},
		{
			name: "invalid dept and valid name , email  and designation",
			lecturer: managementsystem.Lecturer{
				Name:        "Akash",
				Email:       "akash@gmail.com",
				Dept:        "",
				Designation: "senior lecturer",
			},
			willpass: false,
		},
		{
			name: "invalid designation and valid name , email , ",
			lecturer: managementsystem.Lecturer{
				Name:        "Akash",
				Email:       "akash@gmail.com",
				Dept:        "CSE",
				Designation: "",
			},
			willpass: false,
		},
		{
			name: "valid name  ,dept, designation and email without prefix",
			lecturer: managementsystem.Lecturer{
				Name:        "Akash",
				Email:       "@gmail.com",
				Dept:        "CSE",
				Designation: "senior lecturer",
			},
			willpass: false,
		},
		{
			name: "valid email ,dept, designation and withspace name ",
			lecturer: managementsystem.Lecturer{
				Name:        "   ",
				Email:       "akash@gmail.com",
				Dept:        "CSE",
				Designation: "senior lecturer",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlinstance.DB.Exec("DELETE FROM lecturers ")
			mysqlinstance.DB.Exec("ALTER TABLE lecturers AUTO_INCREMENT=1")
			redisInstance.Client.FlushAll(context.Background())

			userBytes, err := json.Marshal(tt.lecturer)
			if err != nil {
				log.Panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/lecturers", buffer)
			w := httptest.NewRecorder()

			handler.CreateLecturersHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var lecturers managementsystem.Lecturer
				if err := json.NewDecoder(w.Body).Decode(&lecturers); err != nil {
					t.Fatalf("Failed to decode response: %d", err)
				}
				if lecturers.Name != tt.lecturer.Name {
					t.Fatalf("Expected name %s , got %s", tt.lecturer.Name, lecturers.Name)
				}
				if lecturers.Email != tt.lecturer.Email {
					t.Fatalf("Expected email %s, got %s", tt.lecturer.Email, lecturers.Email)
				}
				if lecturers.Dept != tt.lecturer.Dept {
					t.Fatalf("Expected dept %s, got %s", tt.lecturer.Dept, lecturers.Dept)
				}
				if lecturers.Designation != tt.lecturer.Designation {
					t.Fatalf("Expected year %s, got %s", tt.lecturer.Designation, lecturers.Designation)
				}
				if lecturers.ID == 0 {
					t.Fatalf("Expected non zero ID!")
				}
			} else {
				if w.Code != http.StatusBadRequest {
					t.Fatalf("Expected not ok status, got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_GetLecturersHandler(t *testing.T) {

	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MySQL_DSN", "root:root@tcp(127.0.0.1:3306)/management_sys")

	redisInstance, err := managementsystem.ConnectRedis()
	if err != nil {
		panic(err)
	}
	mysqlinstance, err := managementsystem.ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handler := &managementsystem.HybridHandler5{MySQL: mysqlinstance, Redis: redisInstance, Ctx: context.Background()}

	mysqlinstance.DB.Exec("DELETE FROM lecturers")
	mysqlinstance.DB.Exec("ALTER TABLE lecturers AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO lecturers (name , email, dept , designation) VALUES (?, ?, ?, ?)", "Akash", "akash@gmail.com", "CSE", "senior lecturer")
	if err != nil {
		t.Fatalf("insert fail: %v", err)
	}
	insertedID, _ := res.LastInsertId()

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid id",
			id:       int(insertedID),
			willpass: true,
		},
		{
			name:     "invalid id ",
			id:       6347,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/students/"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()

			handler.GetLecturersHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var lecturers managementsystem.Lecturer
				if err := json.NewDecoder(w.Body).Decode(&lecturers); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if lecturers.ID != tt.id {
					t.Fatalf("Expected id %d, got %d", tt.id, lecturers.ID)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_UpdateLecturersHandler(t *testing.T) {

	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MySQL_DSN", "root:root@tcp(127.0.0.1:3306)/management_sys")

	redisInstance, err := managementsystem.ConnectRedis()
	if err != nil {
		panic(err)
	}
	mysqlinstance, err := managementsystem.ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handler := &managementsystem.HybridHandler5{MySQL: mysqlinstance, Redis: redisInstance, Ctx: context.Background()}

	mysqlinstance.DB.Exec("DELETE FROM lecturers")
	mysqlinstance.DB.Exec("ALTER TABLE lecturers AUTO_INCREMENT = 1")
	redisInstance.Client.FlushAll(context.Background())
	_, err = mysqlinstance.DB.Exec("INSERT INTO lecturers (id , name , email,  dept , designation) VALUES (1 , 'ramesh' , 'ramesh@gmail.com',  'CSE', 'senior lecturer') ON DUPLICATE KEY UPDATE name='ramesh',email='ramesh@gmail.com',dept='CSE',designation='senior lecturer' ")
	if err != nil {
		t.Fatalf("insert fail: %v", err)
	}
	tests := []struct {
		name     string // description of this test case
		lecturer managementsystem.Lecturer
		willpass bool
	}{
		{
			name: "valid update",
			lecturer: managementsystem.Lecturer{
				ID:          1,
				Name:        "ramesh bhadwa",
				Email:       "rameshbhadwa@gmail.com",
				Dept:        "CSE",
				Designation: "s senior lecturer",
			},
			willpass: true,
		},
		{
			name: "invalid ID format",
			lecturer: managementsystem.Lecturer{
				ID:          4367,
				Name:        "Akash",
				Email:       "akash@gmail.com",
				Dept:        "CSE",
				Designation: "senior lecturer",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.lecturer)
			if err != nil {
				log.Panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPut, "/students/"+fmt.Sprint(tt.lecturer.ID), buffer)
			r = mux.SetURLVars(r, map[string]string{"id": fmt.Sprint(tt.lecturer.ID)})
			w := httptest.NewRecorder()

			handler.UpdateLecturersHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status, got %d", w.Code)
				}
				var updated managementsystem.Lecturer
				if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if updated.Name != tt.lecturer.Name {
					t.Fatalf(" Expected name %s, got %s", tt.lecturer.Name, updated.Name)
				}
				if updated.Email != tt.lecturer.Email {
					t.Fatalf(" Expected email %s, got %s", tt.lecturer.Email, updated.Email)
				}
				if updated.Dept != tt.lecturer.Dept {
					t.Fatalf(" Expected dept %s, got %s", tt.lecturer.Dept, updated.Dept)
				}
				if updated.Designation != tt.lecturer.Designation {
					t.Fatalf(" Expected designation %s, got %s", tt.lecturer.Designation, updated.Designation)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_DeleteLecturersHandler3(t *testing.T) {

	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MySQL_DSN", "root:root@tcp(127.0.0.1:3306)/management_sys")

	redisInstance, err := managementsystem.ConnectRedis()
	if err != nil {
		panic(err)
	}
	mysqlinstance, err := managementsystem.ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handler := &managementsystem.HybridHandler5{MySQL: mysqlinstance, Redis: redisInstance, Ctx: context.Background()}

	mysqlinstance.DB.Exec("DELETE FROM lecturers")
	mysqlinstance.DB.Exec("ALTER TABLE lecturers AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO lecturers (id , name , email, dept , designation) VALUES (?, ?, ?, ?, ?)", 1, "ramesh", "ramesh@gmail.com", "CSE", "senior lecturer")
	if err != nil {
		t.Fatalf("insert fail: %v", err)
	}
	insertedID, _ := res.LastInsertId()

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid id",
			id:       int(insertedID),
			willpass: true,
		},
		{
			name:     "invalid id",
			id:       6356,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodDelete, "/students/"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()

			handler.DeleteLecturersHandler3(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status, got %d", w.Code)
				}
				if w.Body.String() != "lecturer deleted" {
					t.Errorf("Expected 'lecturer deleted', got %s", w.Body.String())
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}
