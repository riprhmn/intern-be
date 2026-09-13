package models

import "time"

type TransactionHeader struct {
	ID           uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	DocNumber    string              `gorm:"column:doc_number;type:varchar(100);not null;unique" json:"doc_number"`
	ApprovalType string              `gorm:"column:approval_type;type:varchar(10);not null" json:"approval_type"` // CN / ANCR
	SectionCode  string              `gorm:"column:section_code;type:varchar(100);not null" json:"section_code"`
	CurrentSeq   int                 `gorm:"column:current_seq;not null;default:1" json:"current_seq"`
	Status       string              `gorm:"column:status;type:varchar(50);not null;default:'IN PROGRESS'" json:"status"` // IN PROGRESS, APPROVED, REJECTED
	CreatedByID  uint64              `gorm:"column:created_by;not null" json:"created_by_id"`
	CreatedAt    time.Time           `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time           `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy    User                `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Details      []TransactionDetail `gorm:"foreignKey:HeaderID" json:"details,omitempty"`
}

func (TransactionHeader) TableName() string {
	return "magang.transaction_headers"
}

type TransactionDetail struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	HeaderID             uint64    `gorm:"column:header_id;not null;index" json:"header_id"`
	SeqStage             int       `gorm:"column:seq_stage;not null" json:"seq_stage"`
	TargetCodeName       string    `gorm:"column:target_code_name;type:varchar(50);not null" json:"target_code_name"`
	ExecutedByUserID     *uint64   `gorm:"column:executed_by_user_id" json:"executed_by_user_id"`
	DelegatedFromUserID *uint64   `gorm:"column:delegated_from_user_id" json:"delegated_from_user_id"`
	Action               string    `gorm:"column:action;type:varchar(20)" json:"action"` // APPROVE, REJECT
	Notes                string    `gorm:"column:notes;type:text" json:"notes"`
	ExecutedAt           *time.Time `gorm:"column:executed_at" json:"executed_at"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ExecutedByUser       *User     `gorm:"foreignKey:ExecutedByUserID" json:"executed_by_user,omitempty"`
	DelegatedFromUser   *User     `gorm:"foreignKey:DelegatedFromUserID" json:"delegated_from_user,omitempty"`
}

func (TransactionDetail) TableName() string {
	return "magang.transaction_details"
}
