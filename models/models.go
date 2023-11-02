package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Order struct {
	ID                string    `json:"-" gorm:"unique;primaryKey"`
	OrderUID          string    `json:"order_uid"`
	DeliveryID        string    `json:"-"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery" gorm:"foreignKey:DeliveryID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Payment           Payment   `json:"payment" gorm:"foreignKey:OrderUID;references:Transaction;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Items             []Item    `json:"items" gorm:"foreignKey:TrackNumber;references:TrackNumber;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	UID     string `json:"-" gorm:"type:uuid;unique;primaryKey"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction" gorm:"unique;primaryKey"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" gorm:"unique;primaryKey"`
	TrackNumber string `json:"track_number" gorm:"not null"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func (deli *Delivery) BeforeCreate(tx *gorm.DB) (err error) {
	deli.UID = uuid.NewString()
	return
}

func (ord *Order) BeforeCreate(tx *gorm.DB) (err error) {
	ord.ID = uuid.NewString()
	return
}
