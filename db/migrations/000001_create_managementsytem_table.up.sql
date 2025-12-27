USE management_sys;

CREATE TABLE IF NOT EXISTS students(
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    age INT NOT NULL,
    dept VARCHAR(50),
    year INT NOT NULL
);

CREATE TABLE IF NOT EXISTS lecturers(
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    dept VARCHAR(50),
    designation VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS books(
    book_id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    author VARCHAR(100) UNIQUE,
    available_copies INT
);
CREATE TABLE IF NOT EXISTS borrow_records(
    borrow_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    user_type ENUM('students','lecturers'),
    book_id INT NOT NULL,
    borrow_date DATE,
    return_date DATE,
    FOREIGN KEY (book_id) REFERENCES books(book_id)
);