package models

import "time"

type ChangeNote struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"                 json:"id"`
	UserID           uint64    `gorm:"column:user_id;not null"                  json:"user_id"`
	Initiator        string    `gorm:"column:initiator;type:varchar(255);not null" json:"initiator"`
	EmpIDInitiator   string    `gorm:"column:emp_id_initiator;type:varchar(50)" json:"emp_id_initiator"`
	Department       string    `gorm:"column:department;type:varchar(100)"      json:"department"`
	ChangePertainsTo string    `gorm:"column:change_pertains_to;type:text;not null" json:"change_pertains_to"`
	Status           string    `gorm:"column:status;type:varchar(50);not null;default:'SUBMITTED'" json:"status"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"         json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"         json:"updated_at"`
}

func (ChangeNote) TableName() string {
	return "magang.change_notes"
}
