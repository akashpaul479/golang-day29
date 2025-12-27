package managementsystem_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	managementsystem "managementsystem/managementsystem"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestHybridHandler5_CreateStudentsHandler(t *testing.T) {

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
		student  managementsystem.Student
		willpass bool
	}{
		{
			name: "valide",
			student: managementsystem.Student{
				Name:  "Akash",
				Email: "akash@gmail.com",
				Age:   21,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: true,
		},
		{
			name: "invalid name and valid email , age , dept, and year",
			student: managementsystem.Student{
				Name:  "",
				Email: "akash@gmail.com",
				Age:   21,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: false,
		},
		{
			name: "invalid email and valid name , age , dept and year",
			student: managementsystem.Student{
				Name:  "Akash",
				Email: "",
				Age:   21,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: false,
		},
		{
			name: "invalid age and valid name , email , dept and year",
			student: managementsystem.Student{
				Name:  "Akash",
				Email: "akash@gmail.com",
				Age:   -1,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: false,
		},
		{
			name: "invalid dept and valid name , email , age and year",
			student: managementsystem.Student{
				Name:  "Akash",
				Email: "akash@gmail.com",
				Age:   20,
				Dept:  "",
				Year:  3,
			},
			willpass: false,
		},
		{
			name: "invalid year and valid name , email , age and dept",
			student: managementsystem.Student{
				Name:  "Akash",
				Email: "akash@gmail.com",
				Age:   20,
				Dept:  "CSE",
				Year:  -1,
			},
			willpass: false,
		},
		{
			name: "valid name , age ,dept, year and email without prefix",
			student: managementsystem.Student{
				Name:  "Akash",
				Email: "@gmail.com",
				Age:   20,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: false,
		},
		{
			name: "valid email , age ,dept, year and withspace name ",
			student: managementsystem.Student{
				Name:  "   ",
				Email: "akash@gmail.com",
				Age:   20,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlinstance.DB.Exec("DELETE FROM students")
			mysqlinstance.DB.Exec("ALTER TABLE students AUTO_INCREMENT = 1")
			redisInstance.Client.FlushAll(context.Background())
			userBytes, err := json.Marshal(tt.student)
			if err != nil {
				log.Panic("Failed to marshal!")
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/students/", buffer)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.CreateStudentsHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var students managementsystem.Student
				if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
					t.Fatalf("Failed to decode response: %d", err)
				}
				if students.Name != tt.student.Name {
					t.Fatalf("Expected name %s , got %s", tt.student.Name, students.Name)
				}
				if students.Email != tt.student.Email {
					t.Fatalf("Expected email %s, got %s", tt.student.Email, students.Email)
				}
				if students.Age != tt.student.Age {
					t.Fatalf("Expected age %d, got %d", tt.student.Age, students.Age)
				}
				if students.Dept != tt.student.Dept {
					t.Fatalf("Expected dept %s, got %s", tt.student.Dept, students.Dept)
				}
				if students.Year != tt.student.Year {
					t.Fatalf("Expected year %d, got %d", tt.student.Year, students.Year)
				}
				if students.ID == 0 {
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

func TestHybridHandler5_GetStudentsHandler(t *testing.T) {

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

	mysqlinstance.DB.Exec("DELETE FROM students")
	mysqlinstance.DB.Exec("ALTER TABLE students AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO students (name , email, age , dept , year) VALUES (?, ?, ?, ?, ?)", "Akash", "akash@gmail.com", 20, "CSE", 3)
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
			id:       5348,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/students/"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()

			handler.GetStudentsHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var students managementsystem.Student
				if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if students.ID != tt.id {
					t.Fatalf("Expected id %d, got %d", tt.id, students.ID)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_UpdatestudentsHandler(t *testing.T) {
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

	mysqlinstance.DB.Exec("DELETE FROM students")
	mysqlinstance.DB.Exec("ALTER TABLE students AUTO_INCREMENT = 1")
	redisInstance.Client.FlushAll(context.Background())
	_, err = mysqlinstance.DB.Exec("INSERT INTO students (id , name , email, age , dept , year) VALUES (1 , 'Akash' , 'akash@gmail.com', 20, 'CSE', 3) ON DUPLICATE KEY UPDATE name='Akash',email='akash@gmail.com',age=20,dept='CSE',year=3 ")
	if err != nil {
		t.Fatalf("insert fail: %v", err)
	}

	tests := []struct {
		name     string // description of this test case
		students managementsystem.Student
		willpass bool
	}{
		{
			name: "valid update",
			students: managementsystem.Student{
				ID:    1,
				Name:  "Akash paul",
				Email: "akashpaul@gmail.com",
				Age:   20,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: true,
		},
		{
			name: "invalid ID format",
			students: managementsystem.Student{
				ID:    4367,
				Name:  "Akash",
				Email: "akash@gmail.com",
				Age:   20,
				Dept:  "CSE",
				Year:  3,
			},
			willpass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.students)
			if err != nil {
				log.Panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPut, "/students/"+fmt.Sprint(tt.students.ID), buffer)
			r = mux.SetURLVars(r, map[string]string{"id": fmt.Sprint(tt.students.ID)})
			w := httptest.NewRecorder()

			handler.UpdatestudentsHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status, got %d", w.Code)
				}
				var updated managementsystem.Student
				if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if updated.Name != tt.students.Name {
					t.Fatalf(" Expected name %s, got %s", tt.students.Name, updated.Name)
				}
				if updated.Email != tt.students.Email {
					t.Fatalf(" Expected email %s, got %s", tt.students.Email, updated.Email)
				}
				if updated.Age != tt.students.Age {
					t.Fatalf(" Expected age %d, got %d", tt.students.Age, updated.Age)
				}
				if updated.Dept != tt.students.Dept {
					t.Fatalf(" Expected dept %s, got %s", tt.students.Dept, updated.Dept)
				}
				if updated.Year != tt.students.Year {
					t.Fatalf(" Expected year %d, got %d", tt.students.Year, updated.Year)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_DeleteStudentsHandler(t *testing.T) {

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

	mysqlinstance.DB.Exec("DELETE FROM students")
	mysqlinstance.DB.Exec("ALTER TABLE students AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO students (id , name , email, age , dept , year) VALUES (? ,?, ?, ?, ?, ?)", 1, "Akash", "akash@gmail.com", 20, "CSE", 3)
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

			handler.DeleteStudentsHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status, got %d", w.Code)
				}
				if w.Body.String() != "student deleted" {
					t.Errorf("Expected 'student deleted', got %s", w.Body.String())
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}
