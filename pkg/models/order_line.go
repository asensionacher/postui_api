package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderLine struct {
	ID        uint            `json:"id" gorm:"primary_key"`
	ProductID uint            `json:"product_id"`
	Quantity  decimal.Decimal `json:"quantity" gorm:"type:decimal(10,2)"` // decimal.NewFromString("136.02")
	Price     uint16          `json:"price"`                              // In Cents, with VAT
	Vat       uint16          `json:"vat"`                                // (ex: 2100 for 21.00%)
	Total     uint16          `json:"total"`                              // In Cents, with VAT
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateOrderLine struct {
	ProductID uint            `json:"product_id" binding:"required"`
	Quantity  decimal.Decimal `json:"quantity" gorm:"type:decimal(10,2)" binding:"required"` // decimal.NewFromString("136.02")
	Price     uint16          `json:"price" binding:"required"`                              // In Cents, with VAT
	Vat       uint16          `json:"vat" binding:"required"`                                // (ex: 2100 for 21.00%)
	Total     uint16          `json:"total" binding:"required"`                              // In Cents
}

type UpdateOrderLine struct {
	ProductID uint            `json:"product_id"`
	Quantity  decimal.Decimal `json:"quantity" gorm:"type:decimal(10,2)"` // decimal.NewFromString("136.02")
	Price     uint16          `json:"price"`                              // In Cents, with VAT
	Vat       uint16          `json:"vat"`                                // (ex: 2100 for 21.00%)
	Total     uint16          `json:"total"`                              // In Cents, with VAT
}
