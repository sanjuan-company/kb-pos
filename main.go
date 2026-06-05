package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type App struct {
	db *sql.DB
}

type UserRequest struct {
	FullName         string `json:"fullname"`
	FirstName        string `json:"fname"`
	LastName         string `json:"lname"`
	MiddleName       string `json:"mname"`
	StaffID          string `json:"staffid"`
	ContactNumber    string `json:"contactnumber"`
	BirthDate        string `json:"birthdate"`
	Email            string `json:"email"`
	TelephoneNumber  string `json:"telephonenumber"`
	MainAddress      string `json:"mainaddress"`
	SecondaryAddress string `json:"secondaryaddress"`
	LastAddress      string `json:"lastaddress"`
	UserRoleID       int    `json:"userroleid"`
	Password         string `json:"userpassword"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	app := &App{db: db}

	r := gin.Default()
	r.POST("/addusers", app.addUserHandler)
	r.PUT("/updateuser/:id", app.updateUserHandler)
	r.DELETE("/deleteuser/:id", app.deleteUserHandler)
	r.GET("/health", healthHandler)

	port := envOrDefault("APP_PORT", "8080")
	log.Printf("starting server on :%s", port)
	log.Fatal(r.Run(":" + port))
}

func openDB() (*sql.DB, error) {
	user := envOrDefault("POSTGRES_USER", "myuser")
	password := envOrDefault("POSTGRES_PASSWORD", "8013075")
	dbName := envOrDefault("POSTGRES_DB", "postgres")
	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "5433")

	if user == "" || password == "" || dbName == "" {
		return nil, fmt.Errorf("missing required environment variables: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (a *App) addUserHandler(c *gin.Context) {
	var req UserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{Success: false, Message: "invalid request body"})
		return
	}

	if err := a.addUser(req); err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ApiResponse{Success: true, Message: "user added successfully"})
}

func (a *App) updateUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{Success: false, Message: "invalid user id"})
		return
	}

	var req UserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{Success: false, Message: "invalid request body"})
		return
	}

	if err := a.updateUser(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{Success: true, Message: "user updated successfully"})
}

func (a *App) deleteUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{Success: false, Message: "invalid user id"})
		return
	}

	if err := a.deleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{Success: true, Message: "user deleted successfully"})
}

func (a *App) addUser(req UserRequest) error {
	_, err := a.db.Exec(
		"SELECT AddUser($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		req.FullName,
		req.FirstName,
		req.LastName,
		req.MiddleName,
		req.StaffID,
		req.ContactNumber,
		req.Email,
		req.TelephoneNumber,
		req.MainAddress,
		req.SecondaryAddress,
		req.LastAddress,
		req.BirthDate,
		req.UserRoleID,
		req.Password,
	)
	return err
}

func (a *App) updateUser(id int64, req UserRequest) error {
	_, err := a.db.Exec(
		"SELECT updateuser($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		id,
		req.FullName,
		req.FirstName,
		req.LastName,
		req.MiddleName,
		req.StaffID,
		req.ContactNumber,
		req.BirthDate,
		req.Email,
		req.UserRoleID,
	)
	return err
}

func (a *App) deleteUser(id int64) error {
	_, err := a.db.Exec("SELECT deleteuser($1)", id)
	return err
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, ApiResponse{Success: true, Message: "ok"})
}

func envOrDefault(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}
