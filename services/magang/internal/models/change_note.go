package models

import "time"

type ApprovalStage struct {
	Label      string     `json:"label"`
	UserID     uint64     `json:"user_id"`
	DelegateID uint64     `json:"delegate_id"`
	Status     string     `json:"status"`
	ActorID    uint64     `json:"actor_id"`
	Comment    string     `json:"comment"`
	ActedAt    *time.Time `json:"acted_at"`
}
type CNEvent struct {
	Action    string    `json:"action"`
	ActorID   uint64    `json:"actor_id"`
	ActorName string    `json:"actor_name"`
	Comment   string    `json:"comment"`
	At        time.Time `json:"at"`
}
type CNNumberCounter struct {
	Kind       string `gorm:"primaryKey;type:varchar(32)"`
	Year       int    `gorm:"primaryKey"`
	LastNumber uint64 `gorm:"not null;default:0"`
}

func (CNNumberCounter) TableName() string { return "magang.cn_number_counters" }

type ChangeNote struct {
	EffectedDate          string          `json:"effected_date"`
	TWA                   string          `json:"twa"`
	ProposedChange        string          `gorm:"type:text" json:"proposed_change"`
	ReasonForChange       string          `gorm:"type:text" json:"reason_for_change"`
	SupportingDocuments   string          `gorm:"type:text" json:"supporting_documents"`
	RevisedDocuments      []CNDocument    `gorm:"serializer:json;type:text" json:"revised_documents"`
	RelatedDocuments      []CNDocument    `gorm:"serializer:json;type:text" json:"related_documents"`
	DocumentCode          string          `gorm:"index" json:"document_code"`
	Section               string          `json:"section"`
	ProcessDescription    string          `gorm:"type:text" json:"process_description"`
	WorkflowSource        string          `gorm:"column:workflow_source;type:varchar(30);not null;default:'PENDING_MASTER_MAPPING'" json:"workflow_source"`
	WorkflowSection       string          `gorm:"column:workflow_section;type:varchar(100)" json:"workflow_section"`
	ProcessedDocumentName string          `json:"processed_document_name"`
	ProcessedDocumentData string          `gorm:"type:text" json:"processed_document_data"`
	IsPIC                 uint64          `gorm:"column:is_pic" json:"is_pic"`
	FinalResultDocument   string          `gorm:"column:final_result_document;type:text" json:"final_result_document"`
	Stages                []ApprovalStage `gorm:"serializer:json;type:text" json:"stages"`
	History               []CNEvent       `gorm:"serializer:json;type:text" json:"history"`
	Version               uint64          `gorm:"not null;default:0" json:"version"`
	RegistrationNumber    string          `gorm:"column:registration_number;type:varchar(50)" json:"registration_number"`

	ID               uint64    `gorm:"primaryKey;autoIncrement"                 json:"id"`
	UserID           uint64    `gorm:"column:user_id;not null"                  json:"user_id"`
	Initiator        string    `gorm:"column:initiator;type:varchar(255);not null" json:"initiator"`
	EmpIDInitiator   string    `gorm:"column:emp_id_initiator;type:varchar(50)" json:"emp_id_initiator"`
	Department       string    `gorm:"column:department;type:varchar(100)"      json:"department"`
	ChangePertainsTo string    `gorm:"column:change_pertains_to;type:text;not null" json:"change_pertains_to"`
	DocumentName     string    `gorm:"column:document_name;type:varchar(255)"   json:"document_name"`
	DocumentData     string    `gorm:"column:document_data;type:text"           json:"document_data"`
	Status           string    `gorm:"column:status;type:varchar(50);not null;default:'SUBMITTED'" json:"status"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"         json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"         json:"updated_at"`
}

type CNDocument struct {
	Name               string `json:"name"`
	Data               string `json:"data"`
	DocumentToRevise   string `json:"document_to_revise"`
	RevisionVersion    string `json:"revision_version"`
	RegistrationNumber string `json:"registration_number"`
	Department         string `json:"department"`
}

func (ChangeNote) TableName() string {
	return "magang.change_notes"
}
