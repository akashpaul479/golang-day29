package managementsystem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Lecturer struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Dept        string `json:"dept"`
	Designation string `json:"designation"`
}

// validation
func Validatelecturer(lecturer Lecturer) error {
	if lecturer.Email == "" {
		return fmt.Errorf("email is invalid and empty")
	}
	if strings.TrimSpace(lecturer.Name) == "" {
		return fmt.Errorf("name is invalid and empty")
	}
	if !strings.HasSuffix(lecturer.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and does not contain @gmail.com")
	}
	prefix := strings.TrimSuffix(lecturer.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a prefix before the @gmail.com ")
	}
	if lecturer.Dept == "" {
		return fmt.Errorf("Dept is invalid!")
	}
	if lecturer.Designation == "" {
		return fmt.Errorf("Year is invalid, please enter a valid year")
	}
	return nil
}

// create students
func (h *HybridHandler5) CreateLecturersHandler(w http.ResponseWriter, r *http.Request) {
	var lecturers Lecturer
	if err := json.NewDecoder(r.Body).Decode(&lecturers); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := Validatelecturer(lecturers); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := h.MySQL.DB.Exec("INSERT INTO lecturers (name , email , dept , designation) VALUES (? , ? , ? , ? )", lecturers.Name, lecturers.Email, lecturers.Dept, lecturers.Designation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lecturers.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lecturers)
}

// Get students
func (h *HybridHandler5) GetLecturersHandler(w http.ResponseWriter, r *http.Request) {
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
	row := h.MySQL.DB.QueryRow("SELECT id , name , email , dept , designation FROM lecturers WHERE  id=?", id)

	var lecturers Lecturer
	if err := row.Scan(&lecturers.ID, &lecturers.Name, &lecturers.Email, &lecturers.Dept, &lecturers.Designation); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	jsondata, err := json.Marshal(lecturers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Second)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}

// update  students
func (h *HybridHandler5) UpdateLecturersHandler(w http.ResponseWriter, r *http.Request) {
	var lecturers Lecturer
	if err := json.NewDecoder(r.Body).Decode(&lecturers); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := Validatelecturer(lecturers); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := h.MySQL.DB.Exec("UPDATE lecturers SET name=?,email=?,dept=?,designation=? WHERE id=?", lecturers.Name, lecturers.Email, lecturers.Dept, lecturers.Designation, lecturers.ID)
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
	jsonData, err := json.Marshal(lecturers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	h.Redis.Client.Set(h.Ctx, fmt.Sprint(lecturers.ID), jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Delete student
func (h *HybridHandler5) DeleteLecturersHandler3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)

	res, err := h.MySQL.DB.Exec("DELETE FROM lecturers WHERE id=?", idInt)
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
		http.Error(w, "lecturer not found", http.StatusNotFound)
		return
	}

	h.Redis.Client.Del(h.Ctx, id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("lecturer deleted"))
}
