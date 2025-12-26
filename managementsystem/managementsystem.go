package managementsystem

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type MySQLInstance5 struct {
	DB *sql.DB
}
type RedisInstance5 struct {
	Client *redis.Client
}
type HybridHandler5 struct {
	MySQL *MySQLInstance5
	Redis *RedisInstance5
	Ctx   context.Context
}

func ConnectMySQL() (*MySQLInstance5, error) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_DB"))
	if err != nil {
		return nil, err
	}
	return &MySQLInstance5{DB: db}, nil
}
func ConnectRedis() (*RedisInstance5, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})
	return &RedisInstance5{Client: rdb}, nil
}
func Managementsystem() {
	godotenv.Load()

	redisInstance, err := ConnectRedis()
	if err != nil {
		panic(err)
	}
	mysqlinstance, err := ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handler := HybridHandler5{MySQL: mysqlinstance, Redis: redisInstance, Ctx: context.Background()}

	r := mux.NewRouter()
	// for students
	r.HandleFunc("/students", handler.CreateStudentsHandler).Methods("POST")
	r.HandleFunc("/students/{id}", handler.GetStudentsHandler).Methods("GET")
	r.HandleFunc("/students/{id}", handler.UpdatestudentsHandler).Methods("PUT")
	r.HandleFunc("/students/{id}", handler.DeleteStudentsHandler).Methods("DELETE")

	// for lecturers
	r.HandleFunc("/lecturers", handler.CreateLecturersHandler).Methods("POST")
	r.HandleFunc("/lecturers/{id}", handler.GetLecturersHandler).Methods("GET")
	r.HandleFunc("/lecturers/{id}", handler.UpdateLecturersHandler).Methods("PUT")
	r.HandleFunc("/lecturers/{id}", handler.DeleteLecturersHandler3).Methods("DELETE")

	// for library
	r.HandleFunc("/books", handler.CreateBookHandler).Methods("POST")
	r.HandleFunc("/books", handler.GetBookHandler).Methods("GET")
	r.HandleFunc("/borrow", handler.BorrowBook).Methods("POST")
	r.HandleFunc("/return", handler.ReturnBook).Methods("POST")

	fmt.Println("Server running on port :8080")
	http.ListenAndServe(":8080", r)
}
