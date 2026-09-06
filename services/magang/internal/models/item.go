package models

import "time"

type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "ACTIVE"
	ItemStatusInactive ItemStatus = "INACTIVE"
)

type Item struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement"          json:"id"`
	Name      string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Status    ItemStatus `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'" json:"status"`
	Note      *string    `gorm:"column:note;type:text"             json:"note,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"  json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"  json:"updated_at"`
}

func (Item) TableName() string {
	return "magang.items"
}
