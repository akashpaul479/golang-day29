package managementsystem_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	managementsystem "managementsystem/managementsystem"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestHybridHandler5_CreateBookHandler(t *testing.T) {

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
		library  managementsystem.Book
		willpass bool
	}{
		{
			name: "valid",
			library: managementsystem.Book{
				Title:            "comics",
				Author:           "muzz2",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name: "invalid book_id and valid title,author and available_copies",
			library: managementsystem.Book{
				Book_id:          -1,
				Title:            "comics",
				Author:           "muzz3",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid title and valid book_id, author , and available_copies",
			library: managementsystem.Book{
				Book_id:          1,
				Title:            "",
				Author:           "kunal",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid author and valid book_id , title and available_copies",
			library: managementsystem.Book{
				Book_id:          1,
				Title:            "comics",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "no available_copies and valid book_id , title , author",
			library: managementsystem.Book{
				Book_id:          1,
				Title:            "comics",
				Author:           "kunal",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name: "withspace title and valid book_id, author , and available_copies",
			library: managementsystem.Book{
				Book_id:          1,
				Title:            "   ",
				Author:           "kunal",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "withspace author and valid book_id , title and available_copies",
			library: managementsystem.Book{
				Book_id:          1,
				Title:            "comics",
				Author:           "   ",
				Available_copies: 10,
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlinstance.DB.Exec("DELETE FROM books")
			mysqlinstance.DB.Exec("ALTER TABLE books AUTO_INCREMENT=1")
			redisInstance.Client.FlushAll(context.Background())

			userBytes, err := json.Marshal(tt.library)
			if err != nil {
				log.Panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/book", buffer)
			w := httptest.NewRecorder()

			handler.CreateBookHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("Expected created status , got %d", w.Code)
				}
				var books managementsystem.Book
				if err := json.NewDecoder(w.Body).Decode(&books); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if books.Title != tt.library.Title {
					t.Fatalf("Expected title %s , got %s", tt.library.Title, books.Title)
				}
				if books.Author != tt.library.Author {
					t.Fatalf("Expected author %s , got %s", tt.library.Author, books.Author)
				}
				if books.Available_copies != tt.library.Available_copies {
					t.Fatalf("Expected Available_copies %d, got %d", tt.library.Available_copies, books.Available_copies)
				}
				if books.Book_id <= 0 {
					t.Fatalf("Expected non zero ID!")
				}
			} else {
				if w.Code != http.StatusBadRequest {
					t.Fatalf("Expected bad request status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_GetBookHandler(t *testing.T) {

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

	mysqlinstance.DB.Exec("DELETE FROM books")
	mysqlinstance.DB.Exec("ALTER TABLE books AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO books (title , author , available_copies) VALUES (?, ?, ?)", "comics", "sujan", 10)
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
			name:     "valid id ",
			id:       int(insertedID),
			willpass: true,
		},
		{
			name:     "invalid id",
			id:       5378,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/books/"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()
			handler.GetBookHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var books managementsystem.Book
				if err := json.NewDecoder(w.Body).Decode(&books); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if books.Book_id != tt.id {
					t.Fatalf("Expected id %d, got %d", tt.id, books.Book_id)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_BorrowBook(t *testing.T) {

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

	mysqlinstance.DB.Exec("DELETE FROM books")
	mysqlinstance.DB.Exec("DELETE FROM borrow_records")
	mysqlinstance.DB.Exec("ALTER TABLE books AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO books(title, author, available_copies) VALUES (?, ?, ?)", "GoLang", "Alice", 2)
	if err != nil {
		log.Panic(err)
	}
	book_id, _ := res.LastInsertId()

	tests := []struct {
		name     string // description of this test case
		body     managementsystem.Borrow_records
		willpass bool
	}{
		{
			name: "valid borrow from students",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "student",
				Book_id:   int(book_id),
			},
			willpass: true,
		},
		{
			name: "valid borrow from lecturers",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "lecturer",
				Book_id:   int(book_id),
			},
			willpass: true,
		},
		{
			name: "invalid user_type",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "",
				Book_id:   int(book_id),
			},
			willpass: false,
		},
		{
			name: "book not found",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "student",
				Book_id:   6354,
			},
			willpass: false,
		},
		{
			name: "book unavailable",
			body: managementsystem.Borrow_records{
				User_id:   1045,
				User_type: "student",
				Book_id:   int(book_id),
			},
			willpass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.body)
			if err != nil {
				log.Panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/borrow", buffer)
			w := httptest.NewRecorder()
			handler.BorrowBook(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("Expected created status , got %d", w.Code)
				}

			} else {
				if w.Code == http.StatusCreated {
					t.Fatalf("Expected failure , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler5_ReturnBook(t *testing.T) {

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

	mysqlinstance.DB.Exec("DELETE FROM books")
	mysqlinstance.DB.Exec("DELETE FROM borrow_records")
	mysqlinstance.DB.Exec("ALTER TABLE books AUTO_INCREMENT=1")
	redisInstance.Client.FlushAll(context.Background())

	res, err := mysqlinstance.DB.Exec("INSERT INTO books(title, author, available_copies) VALUES (?, ?, ?)", "GoLang", "Alice", 2)
	if err != nil {
		log.Panic(err)
	}
	book_id, _ := res.LastInsertId()

	_, err = mysqlinstance.DB.Exec("INSERT INTO borrow_records(user_id, user_type,book_id ,borrow_date, return_date)VALUES (? , ? , ? ,CURDATE(), NULL)", 101, "student", book_id)
	if err != nil {
		t.Fatalf("insert borrow fail: %v", err)
	}

	tests := []struct {
		name     string // description of this test case
		body     managementsystem.Borrow_records
		willpass bool
	}{
		{
			name: "valid return student",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "student",
				Book_id:   int(book_id),
			},
			willpass: true,
		},
		{
			name: "invalid user type",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "",
				Book_id:   int(book_id),
			},
			willpass: false,
		},
		{
			name: "book id not exsists",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "student",
				Book_id:   7635,
			},
			willpass: false,
		},
		{
			name: "Already returned case",
			body: managementsystem.Borrow_records{
				User_id:   101,
				User_type: "student",
				Book_id:   int(book_id),
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.body)
			if err != nil {
				log.Panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/return", buffer)
			w := httptest.NewRecorder()

			handler.ReturnBook(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("Expected created status , got %d", w.Code)
				}
			} else {
				if w.Code == http.StatusCreated {
					t.Fatalf("Expected failure , got %d", w.Code)
				}
			}
		})
	}
}
