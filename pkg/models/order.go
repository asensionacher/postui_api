package models

import (
	"time"
	"github.com/lib/pq"
)

type Order struct {
	ID            uint      `json:"id" gorm:"primary_key"`
	Vendor        string    `json:"customer"`
	Total         uint16    `json:"total"` // In cents, with VAT
	LinesID       pq.Int64Array    `json:"lines_id" gorm:"type:bigint[]"`
	CashoutNumber uint      `json:"cashout_number"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateOrder struct {
	Vendor        string `json:"customer" binding:"required"`
	Total         uint16 `json:"total" binding:"required"`                         // In cents, with VAT
	LinesID       pq.Int64Array `json:"lines_id" binding:"required" gorm:"type:bigint[]"` // PostgreSQL array
	CashoutNumber uint   `json:"cashout_number" binding:"required"`
}

type UpdateOrder struct {
	Vendor        string `json:"customer"`
	Total         uint16 `json:"total"`                                            // In cents, with VAT
	LinesID       pq.Int64Array `json:"lines_id" binding:"required" gorm:"type:bigint[]"` // PostgreSQL array
	CashoutNumber uint   `json:"cashout_number"`
}
