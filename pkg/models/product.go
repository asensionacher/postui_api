package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Product struct {
	ID            uint            `json:"id" gorm:"primary_key"`
	Name          string          `json:"name"`
	Price         uint16          `json:"price"`                           // In cents, with VAT
	Vat           uint16          `json:"vat"`                             // (ex: 2100 for 21.00%)
	Stock         decimal.Decimal `json:"stock" gorm:"type:decimal(10,2)"` // decimal.NewFromString("136.02")
	BarcodeNumber string          `json:"barcode_number"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateProducts struct {
	Name          string          `json:"name" binding:"required"`
	Price         uint16          `json:"price" binding:"required"` // In cents, with VAT
	Vat           uint16          `json:"vat" binding:"required"`   // (ex: 2100 for 21.00%)
	Stock         decimal.Decimal `json:"stock" gorm:"type:decimal(10,2)" binding:"required"`
	BarcodeNumber string          `json:"barcode_number" binding:"required"`
}

type UpdateProduct struct {
	Name          string          `json:"name"`
	Price         uint16          `json:"price"` // In cents, with VAT
	Vat           uint16          `json:"vat"`   // (ex: 2100 for 21.00%)
	Stock         decimal.Decimal `json:"stock" gorm:"type:decimal(10,2)"`
	BarcodeNumber string          `json:"barcode_number"`
}
