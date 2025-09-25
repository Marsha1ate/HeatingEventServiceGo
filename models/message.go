// models/Message.go
package models

import (
    "time"

    "gorm.io/gorm"
)

var DB *gorm.DB

type Message struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    Message    string    `gorm:"not null" json:"message"`
    Severity   int       `gorm:"not null" json:"severity"`
    Source     string    `gorm:"not null" json:"source"`
    // Измените gorm-теги для соответствия с именами столбцов в БД
    ServerTime time.Time `gorm:"column:servertime;default:current_timestamp" json:"serverTime"`
    SourceTime time.Time `gorm:"column:sourcetime;default:current_timestamp" json:"sourceTime"`
}

func (Message) TableName() string {
    return "messages"
}