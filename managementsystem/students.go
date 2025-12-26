package managementsystem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Student struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
	Dept  string `json:"dept"`
	Year  int    `json:"year"`
}

// validation
func ValidateUser(student Student) error {
	if student.Email == "" {
		return fmt.Errorf("email is invalid and empty")
	}
	if strings.TrimSpace(student.Name) == "" {
		return fmt.Errorf("name is invalid and empty")
	}
	if !strings.HasSuffix(student.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and does not contain @gmail.com")
	}
	prefix := strings.TrimSuffix(student.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a prefix before the @gmail.com ")
	}
	if student.Age <= 0 {
		return fmt.Errorf("Age is less than 0")
	}
	if student.Age >= 100 {
		return fmt.Errorf("Age is grater than 0")
	}
	if student.Dept == "" {
		return fmt.Errorf("Dept is invalid!")
	}
	if student.Year == 0 {
		return fmt.Errorf("Year is invalid, please enter a valid year")
	}
	return nil
}

// create students
func (h *HybridHandler5) CreateStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var students Student
	if err := json.NewDecoder(r.Body).Decode(&students); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateUser(students); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := h.MySQL.DB.Exec("INSERT INTO students (name , email, age , dept , year) VALUES (? , ? , ? , ? , ?)", students.Name, students.Email, students.Age, students.Dept, students.Year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	students.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(students)
}

// Get students
func (h *HybridHandler5) GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	value, err := h.Redis.Client.Get(h.Ctx, id).Result()
	if err == nil {
		log.Println("Cache hit!")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}
	fmt.Println("Cache miss Quering MySQL ...")
	row := h.MySQL.DB.QueryRow("SELECT id , name , email , age , dept , year FROM students WHERE  id=?", id)

	var students Student
	if err := row.Scan(&students.ID, &students.Name, &students.Email, &students.Age, &students.Dept, &students.Year); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	jsondata, err := json.Marshal(students)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Second)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}

// update  students
func (h *HybridHandler5) UpdatestudentsHandler(w http.ResponseWriter, r *http.Request) {
	var students Student
	if err := json.NewDecoder(r.Body).Decode(&students); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateUser(students); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := h.MySQL.DB.Exec("UPDATE students SET name=?,email=?,age=?,dept=?,year=? WHERE id=?", students.Name, students.Email, students.Age, students.Dept, students.Year, students.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	jsonData, err := json.Marshal(students)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	h.Redis.Client.Set(h.Ctx, fmt.Sprint(students.ID), jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Delete student
func (h *HybridHandler5) DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)

	res, err := h.MySQL.DB.Exec("DELETE FROM students WHERE id=?", idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	h.Redis.Client.Del(h.Ctx, id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("user deleted"))
}
