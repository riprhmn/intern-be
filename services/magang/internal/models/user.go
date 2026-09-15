package models

import "time"

type User struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement"                               json:"id"`
	NoEmp            string     `gorm:"column:no_emp;type:varchar(50)"                         json:"no_emp"`
	IdEmp            string     `gorm:"column:id_emp;type:varchar(50)"                         json:"id_emp"`
	CodeName         string     `gorm:"column:code_name;type:varchar(50)"                     json:"code_name"`
	LevelRank        int        `gorm:"column:level_rank;type:integer;default:4"               json:"level_rank"`
	Username         string     `gorm:"column:username;type:varchar(100);not null;unique"       json:"username"`
	Password         string     `gorm:"column:password;type:varchar(255);not null"             json:"-"`
	FullName         string     `gorm:"column:full_name;type:varchar(255);not null"            json:"full_name"`
	Email            string     `gorm:"column:email;type:varchar(255)"                        json:"email"`
	Title            string     `gorm:"column:title;type:varchar(150)"                        json:"title"`
	Company          string     `gorm:"column:company;type:varchar(255)"                      json:"company"`
	Role             string     `gorm:"column:role;type:varchar(50);not null;default:'user'"   json:"role"`
	Section          string     `gorm:"column:section;type:varchar(100)"                       json:"section"`
	Department       string     `gorm:"column:department;type:varchar(100)"                    json:"department"`
	Division         string     `gorm:"column:division;type:varchar(100)"                      json:"division"`
	IsActive         bool       `gorm:"column:is_active;not null;default:true"                 json:"is_active"`
	ResetToken       string     `gorm:"column:reset_token;type:varchar(255)"                   json:"-"`
	ResetTokenExpiry *time.Time `gorm:"column:reset_token_expiry"                             json:"-"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"                       json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"                       json:"updated_at"`
}

func (User) TableName() string {
	return "magang.users"
}
