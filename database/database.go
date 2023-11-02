package database

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"streaming-project/models"
)

type Database struct {
	dbName      string
	dbUser      string
	dbPassword  string
	dbPort      int
	dbInterface *gorm.DB
	dbCache     *OrderCache
}

func NewDatabase(user string, password string, databaseName string, port int) (*Database, error) {

	postgresConfig := fmt.Sprintf("user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		user, password, databaseName, port)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  postgresConfig,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return new(Database), err
	}

	cache := NewCache()

	database := &Database{dbName: databaseName, dbUser: user, dbPassword: password, dbPort: port, dbInterface: gormDB, dbCache: cache}

	return database, err
}

func (db *Database) Init() error {
	err := db.dbInterface.AutoMigrate(&models.Delivery{}, &models.Payment{}, &models.Item{})
	if err != nil {
		return err
	}

	err = db.dbInterface.AutoMigrate(&models.Order{})
	if err != nil {
		return err
	}

	orders, err := db.getAllOrders()
	if err != nil {
		return err
	}

	for _, order := range orders {
		db.dbCache.Update(order.OrderUID, order)
	}

	return nil
}

func (db *Database) AddOrder(order *models.Order) error {
	err := db.dbInterface.First(order, "order_uid = ?", order.OrderUID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Order not found in the database, creating new record")
		result := db.dbInterface.Create(order)
		if result.Error != nil {
			log.Println(result.Error, "continuing")
			return nil
		}
		db.dbCache.Update(order.OrderUID, *order)
		return nil

	} else {
		return err
	}
}

func (db *Database) getAllOrders() ([]models.Order, error) {
	var orders []models.Order
	result := db.dbInterface.Preload("Items").Preload("Delivery").Preload("Payment").Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}
	fmt.Println(result)
	return orders, nil
}

func (db *Database) GetOrderByID(id string) (*models.Order, error) {
	result, ok := db.dbCache.Read(id)
	if !ok {
		return nil, errors.New("order not found")
	}
	fmt.Println(result)
	return result, nil
}
