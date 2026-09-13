package models

import "time"

// ApprovalMapping is the reusable approval configuration. It is intentionally
// separate from transaction stages so editing Settings never rewrites an
// approval that is already in progress.
type ApprovalMapping struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ApprovalType string    `gorm:"column:approval_type;type:varchar(10);not null;index:idx_approval_mapping_scope" json:"approval_type"`
	SectionCode  string    `gorm:"column:section_code;type:varchar(100);not null;index:idx_approval_mapping_scope" json:"section_code"`
	Sequence     int       `gorm:"column:sequence;not null" json:"sequence"`
	UserID       uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedBy    uint64    `gorm:"column:created_by;not null" json:"created_by"`
	UpdatedBy    uint64    `gorm:"column:updated_by;not null" json:"updated_by"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
}

func (ApprovalMapping) TableName() string { return "magang.approval_mappings" }

// ApprovalDelegation keeps the mapping owner (FromUserID) unchanged and
// records who may act on their behalf (ToUserID) for a bounded scope.
type ApprovalDelegation struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromUserID   uint64     `gorm:"column:from_user_id;not null;index" json:"from_user_id"`
	ToUserID     uint64     `gorm:"column:to_user_id;not null;index" json:"to_user_id"`
	ApprovalType string     `gorm:"column:approval_type;type:varchar(10);not null;default:'ALL'" json:"approval_type"`
	SectionCode  string     `gorm:"column:section_code;type:varchar(100);not null;default:'ALL'" json:"section_code"`
	StartsAt     *time.Time `gorm:"column:starts_at" json:"starts_at"`
	EndsAt       *time.Time `gorm:"column:ends_at" json:"ends_at"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true;index" json:"is_active"`
	CreatedBy    uint64     `gorm:"column:created_by;not null" json:"created_by"`
	UpdatedBy    uint64     `gorm:"column:updated_by;not null" json:"updated_by"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	FromUser     User       `gorm:"foreignKey:FromUserID" json:"from_user"`
	ToUser       User       `gorm:"foreignKey:ToUserID" json:"to_user"`
}

func (ApprovalDelegation) TableName() string { return "magang.approval_delegations" }
