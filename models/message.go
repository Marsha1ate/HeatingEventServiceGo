package models

import (
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB // This will be set from config

type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Message    string    `gorm:"column:message;not null" json:"message"`
	Severity   int       `gorm:"column:severity;not null" json:"severity"`
	Source     string    `gorm:"column:source;not null" json:"source"`
	ServerTime time.Time `gorm:"column:serverTime;not null;default:current_timestamp" json:"serverTime"`
	SourceTime time.Time `gorm:"column:sourceTime;not null;default:current_timestamp" json:"sourceTime"`
}

func (Message) TableName() string {
	return "messages"
}
