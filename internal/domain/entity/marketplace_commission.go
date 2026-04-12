package entity

import "time"

type MarketplaceCommission struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"                    json:"id"`
	Name       string    `gorm:"type:varchar(100);not null;uniqueIndex"      json:"name"`
	Percentage float64   `gorm:"type:decimal(5,4);not null;default:0.0200"   json:"percentage"`
	WordingES  string    `gorm:"type:text;not null"                          json:"wording_es"`
	WordingEN  string    `gorm:"type:text;not null"                          json:"wording_en"`
	Currency   string    `gorm:"type:varchar(10);not null;default:ARS"       json:"currency"`
	IsActive   bool      `gorm:"not null;default:true"                       json:"is_active"`
	CreatedAt  time.Time `gorm:"autoCreateTime"                              json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"                              json:"updated_at"`
}
