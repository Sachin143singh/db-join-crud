package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderID      uint      `json:"orderId" gorm:"primary_key"`
	CustomerName string    `json:"customerName"`
	OrderedAt    time.Time `json:"orderedAt"`
	Items        []Item    `json:"items" gorm:"foreignkey:OrderID"`
}
type Item struct {
	gorm.Model
	LineItemID  uint   `json:"lineItemId" gorm:"primary_key"`
	ItemCode    string `json:"itemCode"`
	Description string `json:"description"`
	Quantity    uint   `json:"quantity"`
	OrderID     uint   `json:"-"`
}

var db *gorm.DB

func initDB() {
	var err error
	dataSourceName := "root:root@tcp(localhost:3306)/sss?parseTime=True"
	db, err = gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})

	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}

	// Create the database. This is a one-time step.
	// Comment out if running multiple times - You may see an error otherwise
	db.Exec("CREATE DATABASE orders_db")
	db.Exec("USE orders_db")

	// Migration to create tables for Order and Item schema
	db.AutoMigrate(&Order{}, &Item{})
}
func main() {
	router := mux.NewRouter()
	// Create
	router.HandleFunc("/orders", createOrder).Methods("POST")
	// Read
	router.HandleFunc("/orders/{orderId}", getOrder).Methods("GET")
	// Read-all
	router.HandleFunc("/orders", getOrders).Methods("GET")
	// Update
	router.HandleFunc("/orders/{orderId}", updateOrder).Methods("PUT")
	// Delete
	router.HandleFunc("/orders/{orderId}", deleteOrder).Methods("DELETE")
	// Initialize db connection
	initDB()

	log.Fatal(http.ListenAndServe(":6667", router))
}
func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	json.NewDecoder(r.Body).Decode(&order)
	// Creates new order by inserting records in the `orders` and `items` table
	db.Create(&order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
func getOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var orders []Order
	db.Preload("Items").Find(&orders)
	json.NewEncoder(w).Encode(orders)
}
func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	inputOrderID := params["orderId"]

	var order Order
	db.Preload("Items").First(&order, inputOrderID)
	json.NewEncoder(w).Encode(order)
}
func updateOrder(w http.ResponseWriter, r *http.Request) {
	var updatedOrder Order
	json.NewDecoder(r.Body).Decode(&updatedOrder)
	db.Save(&updatedOrder)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}
func deleteOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	inputOrderID := params["orderId"]
	// Convert `orderId` string param to uint64
	id64, _ := strconv.ParseUint(inputOrderID, 10, 64)
	// Convert uint64 to uint
	idToDelete := uint(id64)

	db.Where("order_id = ?", idToDelete).Delete(&Item{})
	db.Where("order_id = ?", idToDelete).Delete(&Order{})
	w.WriteHeader(http.StatusNoContent)
}
