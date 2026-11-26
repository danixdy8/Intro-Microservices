package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Order struct {
	ID             int64   `json:"id"`
	CustomerName   string  `json:"customer_name"`
	ProductSummary string  `json:"product_summary"`
	Quantity       int     `json:"quantity"`
	Total          float64 `json:"total"`
	Status         string  `json:"status"`
	OrderDate      string  `json:"order_date"`
	UpdatedAt      string  `json:"updated_at"`
}

type CreateOrderInput struct {
	CustomerName   string  `json:"customer_name" binding:"required"`
	ProductSummary string  `json:"product_summary" binding:"required"`
	Quantity       int     `json:"quantity" binding:"required,gt=0"`
	Total          float64 `json:"total" binding:"required,gt=0"`
	Status         string  `json:"status"`
}

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

var validStatuses = map[string]bool{
	"PENDING":    true,
	"PROCESSING": true,
	"SHIPPED":    true,
	"DELIVERED":  true,
	"CANCELLED":  true,
}

func initDB(dbPath string) *sql.DB {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("error creando directorio de volumen: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("error abriendo SQLite: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer_name TEXT NOT NULL,
		product_summary TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		total REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'PENDING',
		order_date TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);`

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("error inicializando tabla orders: %v", err)
	}
	return db
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/orders.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	db := initDB(dbPath)
	defer db.Close()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		var count int
		_ = db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
		c.JSON(http.StatusOK, gin.H{
			"service":  "orders-api",
			"language": "Go / Gin",
			"status":   "UP",
			"database": dbPath,
			"records":  count,
		})
	})

	router.POST("/orders", func(c *gin.Context) {
		var input CreateOrderInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de pedido inválidos", "details": err.Error()})
			return
		}
		orderStatus := strings.ToUpper(strings.TrimSpace(input.Status))
		if orderStatus == "" {
			orderStatus = "PENDING"
		}
		if !validStatuses[orderStatus] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Estado inválido. Use PENDING, PROCESSING, SHIPPED, DELIVERED o CANCELLED"})
			return
		}

		res, err := db.Exec(
			"INSERT INTO orders (customer_name, product_summary, quantity, total, status) VALUES (?, ?, ?, ?, ?)",
			strings.TrimSpace(input.CustomerName),
			strings.TrimSpace(input.ProductSummary),
			input.Quantity,
			input.Total,
			orderStatus,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		var created Order
		_ = db.QueryRow("SELECT id, customer_name, product_summary, quantity, total, status, order_date, updated_at FROM orders WHERE id = ?", id).
			Scan(&created.ID, &created.CustomerName, &created.ProductSummary, &created.Quantity, &created.Total, &created.Status, &created.OrderDate, &created.UpdatedAt)

		c.JSON(http.StatusCreated, gin.H{"message": "Pedido creado correctamente", "data": created})
	})

	router.GET("/orders", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, customer_name, product_summary, quantity, total, status, order_date, updated_at FROM orders ORDER BY id ASC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		orders := make([]Order, 0)
		for rows.Next() {
			var o Order
			if err := rows.Scan(&o.ID, &o.CustomerName, &o.ProductSummary, &o.Quantity, &o.Total, &o.Status, &o.OrderDate, &o.UpdatedAt); err == nil {
				orders = append(orders, o)
			}
		}
		c.JSON(http.StatusOK, gin.H{"total": len(orders), "data": orders})
	})

	router.GET("/orders/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pedido inválido"})
			return
		}
		var o Order
		err = db.QueryRow("SELECT id, customer_name, product_summary, quantity, total, status, order_date, updated_at FROM orders WHERE id = ?", id).
			Scan(&o.ID, &o.CustomerName, &o.ProductSummary, &o.Quantity, &o.Total, &o.Status, &o.OrderDate, &o.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pedido no encontrado"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": o})
	})

	updateStatusHandler := func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pedido inválido"})
			return
		}
		var input UpdateStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Debe indicar el campo status"})
			return
		}
		newStatus := strings.ToUpper(strings.TrimSpace(input.Status))
		if !validStatuses[newStatus] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Estado inválido. Use PENDING, PROCESSING, SHIPPED, DELIVERED o CANCELLED"})
			return
		}

		res, err := db.Exec("UPDATE orders SET status = ?, updated_at = datetime('now') WHERE id = ?", newStatus, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pedido no encontrado"})
			return
		}

		var updated Order
		_ = db.QueryRow("SELECT id, customer_name, product_summary, quantity, total, status, order_date, updated_at FROM orders WHERE id = ?", id).
			Scan(&updated.ID, &updated.CustomerName, &updated.ProductSummary, &updated.Quantity, &updated.Total, &updated.Status, &updated.OrderDate, &updated.UpdatedAt)
		c.JSON(http.StatusOK, gin.H{"message": "Estado de pedido actualizado", "data": updated})
	}

	router.PATCH("/orders/:id/status", updateStatusHandler)
	router.PUT("/orders/:id", updateStatusHandler)

	router.DELETE("/orders/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pedido inválido"})
			return
		}
		res, err := db.Exec("DELETE FROM orders WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pedido no encontrado"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Pedido eliminado correctamente", "deleted_id": id})
	})

	log.Printf("Orders API (Go/Gin) escuchando en :%s con SQLite en %s", port, dbPath)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
