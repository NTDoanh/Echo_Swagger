package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "main/docs"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title API Echo_Swagger
// @version 1.0
// @description Nguyen Trong Doanh
// @host localhost:1234
// @BasePath /v2
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	e := echo.New()

	// Cài đặt Echo Swagger
	e.GET("/swagger/*", echoSwagger.EchoWrapHandler())
	e.GET("/v2/users", read)
	e.POST("/v2/users/create", create)
	e.PUT("/v2/users/update/:id", update)
	e.DELETE("/v2/users/delete/:id", delete)
	e.Start(":1234")
}

// API Read (GET): Lấy danh sách người dùng
// @Summary Lấy danh sách người dùng
// @Description Trả về danh sách tất cả người dùng từ cơ sở dữ liệu
// @Tags Users
// @Produce json
// @Success 200 {array} User
// @Router /users [get]
func read(c echo.Context) error {
	connStr := "postgres://postgres:123@localhost:5432/doanhtr?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	// Connect to database with connectionString as DSN
	if err != nil {
		log.Fatal(fmt.Sprintf("fail to open with err: %v", err))
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(fmt.Sprintf("fail to ping read with err: %v", err))
	}
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to query read with err: %v", err))
	}
	defer rows.Close()

	//Slice luu danh sach user
	var users []User
	// Doc du lieu tu truy van vao slice
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			log.Fatal(fmt.Sprintf("failed to scan with err: %v", err))
		}
		users = append(users, user)
	}
	return c.JSON(http.StatusOK, users)
}

// @Summary Tạo người dùng mới
// @Description Tạo một người dùng mới với thông tin được cung cấp
// @Tags Users
// @Accept json
// @Produce json
// @Param user body User true "Thông tin người dùng mới"
// @Success 201 {object} User "Người dùng đã được tạo thành công"
// @Router /users/create [post]
func create(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Kết nối tới cơ sở dữ liệu PostgreSQL
	connStr := "postgres://postgres:123@localhost:5432/doanhtr?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	// Thực hiện truy vấn SQL để tạo người dùng mới
	_, err = db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", user.Name, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}

// API Update (PUT): Cập nhật thông tin người dùng
// @Summary Cập nhật thông tin người dùng
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID của người dùng cần cập nhật"
// @Param user body User true "Thông tin người dùng cần cập nhật"
// @Success 200 {object} User
// @Router /users/update/{id} [put]
func update(c echo.Context) error {

	connStr := "postgres://postgres:123@localhost:5432/doanhtr?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal server error")
	}
	defer db.Close()

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid find id")
	}

	// Giải nén dữ liệu JSON được gửi từ client vào một đối tượng User
	var user User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid Bind")
	}

	// Kết nối vào cơ sở dữ liệu

	defer db.Close()

	// Thực hiện truy vấn SQL để cập nhật thông tin người dùng trong cơ sở dữ liệu
	_, err = db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", user.Name, user.Email, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal query error")
	}

	return c.JSON(http.StatusOK, user)
}

// API Delete (DELETE): Xóa người dùng
// @Summary Xóa người dùng
// @Tags Users
// @Produce json
// @Param id path int true "ID của người dùng cần xóa"
// @Success 200 {string} string "Người dùng đã bị xóa"
// @Router /users/delete/{id} [delete]
func delete(c echo.Context) error {
	connStr := "postgres://postgres:123@localhost:5432/doanhtr?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to open with err: %v", err))
	}
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	defer db.Close()

	var userExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check user existence"})
	}

	if !userExists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	_, err = db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
