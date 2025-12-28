package managementsystem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Book struct {
	Book_id          int    `json:"book_id"`
	Title            string `json:"title"`
	Author           string `json:"author"`
	Available_copies int    `json:"available_copies"`
}
type Borrow_records struct {
	Borrow_id   int        `json:"borrow_id"`
	User_id     int        `json:"user_id"`
	User_type   string     `json:"user_type"`
	Book_id     int        `json:"book_id"`
	Borrow_date time.Time  `json:"borrow_date"`
	Return_date *time.Time `json:"time_date"`
}

// validation
func ValidateLibrary(book Book) error {
	if book.Book_id < 0 {
		return fmt.Errorf("book_id is invalid")
	}
	if strings.TrimSpace(book.Title) == "" {
		return fmt.Errorf("title is invalid and empty")
	}
	if strings.TrimSpace(book.Author) == "" {
		return fmt.Errorf("Author is invalid and empty")
	}
	if book.Available_copies <= 0 {
		return fmt.Errorf("available_copies is less than 0")
	}

	return nil
}

// create students
func (h *HybridHandler5) CreateBookHandler(w http.ResponseWriter, r *http.Request) {
	var books Book
	if err := json.NewDecoder(r.Body).Decode(&books); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateLibrary(books); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := h.MySQL.DB.Exec("INSERT INTO books (title , author , available_copies) VALUES ( ? , ? , ?)", books.Title, books.Author, books.Available_copies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	books.Book_id = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(books)
}

// Get students
func (h *HybridHandler5) GetBookHandler(w http.ResponseWriter, r *http.Request) {
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
	row := h.MySQL.DB.QueryRow("SELECT book_id , title , author , available_copies FROM books WHERE  book_id=?", id)

	var books Book
	if err := row.Scan(&books.Book_id, &books.Title, &books.Author, &books.Available_copies); err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	jsondata, err := json.Marshal(books)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Second)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}

// Borrow books
func (h *HybridHandler5) BorrowBook(w http.ResponseWriter, r *http.Request) {
	var record Borrow_records
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// validate user type
	if record.User_type != "student" && record.User_type != "lecturer" {
		http.Error(w, "invalid user_type, must be 'student' or 'lecturer'", http.StatusBadRequest)
		return
	}
	//  Check if Book is available
	var available int
	err := h.MySQL.DB.QueryRow("SELECT available_copies FROM books WHERE book_id=?", record.Book_id).Scan(&available)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	if available <= 0 {
		http.Error(w, " Book not available", http.StatusBadRequest)
		return
	}
	// Insert borrow record
	_, err = h.MySQL.DB.Exec("INSERT INTO borrow_records(user_id, user_type,book_id ,borrow_date)VALUES (? , ? , ? , CURDATE())", record.User_id, record.User_type, record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//  decrement available copies
	_, err = h.MySQL.DB.Exec("UPDATE books SET available_copies = available_copies-1 WHERE book_id=?", record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Book borrowed"})
}

// Return book
func (h *HybridHandler5) ReturnBook(w http.ResponseWriter, r *http.Request) {
	var record Borrow_records
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "invalid json", http.StatusInternalServerError)
		return
	}
	// validate user type
	if record.User_type != "student" && record.User_type != "lecturer" {
		http.Error(w, "invalid user_type, must be 'student' or 'lecturer'", http.StatusBadRequest)
		return
	}
	// Update borrow_books record with return date
	res, err := h.MySQL.DB.Exec("UPDATE borrow_records SET return_date=CURDATE() WHERE user_id=? AND book_id=? AND return_date IS NULL", record.User_id, record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "no active borrow record found", http.StatusNotFound)
	}
	//  increment available copies
	_, err = h.MySQL.DB.Exec("UPDATE books SET available_copies = available_copies+1 WHERE book_id=?", record.Book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Book return"})

}
