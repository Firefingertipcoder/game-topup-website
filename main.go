package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite" 
)

var db *sql.DB

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Order struct {
	ID        int     `json:"id"`
	GameID    string  `json:"gameId"`
	PackName  string  `json:"packName"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

type TopUpRequest struct {
	GameID string `json:"gameId"`
	Pack   struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	} `json:"pack"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./orders.db")
	if err != nil {
		log.Fatal(err)
	}
	userTable := `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE, email TEXT UNIQUE, password TEXT);`
	orderTable := `CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY AUTOINCREMENT, userId INTEGER, gameId TEXT, packName TEXT, amount REAL, status TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);`
	db.Exec(userTable)
	db.Exec(orderTable)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userId := session.Get("userId")
		if userId == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Please login first!"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	initDB()
	r := gin.Default()
	store := cookie.NewStore([]byte("super-secret-gaming-key"))
	r.Use(sessions.Sessions("mysession", store))

	r.Static("/static", "./public")
	r.StaticFile("/", "./public/home.html")
	r.StaticFile("/login", "./public/login.html")
	r.StaticFile("/register", "./public/register.html")
	r.StaticFile("/shop", "./public/index.html")
	r.StaticFile("/dashboard", "./public/dashboard.html")
    r.StaticFile("/success.html", "./public/success.html")

	r.POST("/api/register", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid data"})
			return
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		_, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", user.Username, user.Email, string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Username or Email already exists!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		c.ShouldBindJSON(&loginData)
		var user User
		err := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", loginData.Username).Scan(&user.ID, &user.Username, &user.Password)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
			return
		}
		session := sessions.Default(c)
		session.Set("userId", user.ID)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	r.GET("/api/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusFound, "/login")
	})

	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())
	{
		authorized.POST("/api/topup", func(c *gin.Context) {
			var req TopUpRequest
			c.ShouldBindJSON(&req)
			session := sessions.Default(c)
			userId := session.Get("userId")
			res, err := db.Exec("INSERT INTO orders (userId, gameId, packName, amount, status) VALUES (?, ?, ?, ?, ?)", 
				userId, req.GameID, req.Pack.Name, req.Pack.Price, "Completed")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "DB Error"})
				return
			}
			id, _ := res.LastInsertId()
			c.JSON(http.StatusOK, gin.H{"success": true, "orderId": id})
		})

		authorized.GET("/api/orders", func(c *gin.Context) {
			session := sessions.Default(c)
			userId := session.Get("userId")
			rows, _ := db.Query("SELECT id, gameId, packName, amount, timestamp FROM orders WHERE userId = ? ORDER BY timestamp DESC", userId)
			defer rows.Close()
			var orders []Order
			for rows.Next() {
				var o Order
				rows.Scan(&o.ID, &o.GameID, &o.PackName, &o.Amount, &o.Timestamp)
				orders = append(orders, o)
			}
			c.JSON(http.StatusOK, orders)
		})
	}

	fmt.Println("🚀 Go Server live at http://localhost:8080")
	r.Run(":8080")
}
