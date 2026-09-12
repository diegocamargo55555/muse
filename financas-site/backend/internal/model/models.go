package model

import "time"

// Account é uma conta de investimento em uma moeda (BRL, USD, EUR...).
type Account struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:120;not null"`
	Currency  string    `json:"currency" gorm:"size:3;not null"`
	CreatedAt time.Time `json:"createdAt"`
}

// Transaction é um lançamento: deposit, withdraw, buy, sell ou dividend.
type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AccountID uint      `json:"accountId" gorm:"index;not null"`
	Type      string    `json:"type" gorm:"size:16;not null"`
	Symbol    string    `json:"symbol" gorm:"size:32"`
	Market    string    `json:"market" gorm:"size:8"`
	Qty       float64   `json:"qty"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency" gorm:"size:3;not null"`
	Date      time.Time `json:"date"`
	Note      string    `json:"note" gorm:"size:255"`
}
