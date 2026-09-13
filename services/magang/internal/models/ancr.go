package models

import "time"

// ANCRRequest menyimpan satu pengajuan Audit Non-Conformance Report
type ANCRRequest struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"                          json:"id"`
	UserID             uint64    `gorm:"column:user_id;not null"                           json:"user_id"`
	RegistrationNumber string    `gorm:"column:registration_number;type:varchar(50)"       json:"registrationNumber"`
	TypeANCR           string    `gorm:"column:type_ancr;type:varchar(20);not null"        json:"typeAncr"`
	AuditNo            string    `gorm:"column:audit_no;type:varchar(100)"                 json:"auditNo"`
	Initiator          string    `gorm:"column:initiator;type:varchar(255)"                json:"initiator"`
	DepartmentInit     string    `gorm:"column:department_initiator;type:varchar(100)"     json:"departmentInitiator"`
	Auditor            string    `gorm:"column:auditor;type:varchar(255)"                  json:"auditor"`
	Auditee            string    `gorm:"column:auditee;type:varchar(255)"                  json:"auditee"`
	DepartmentAuditee  string    `gorm:"column:department_auditee;type:varchar(100)"       json:"departmentAuditee"`
	RelatedTo          string    `gorm:"column:related_to;type:text"                       json:"relatedTo"`
	FindingCriteria    string    `gorm:"column:finding_criteria;type:varchar(50)"          json:"findingCriteria"`
	Problem            string    `gorm:"column:problem;type:text"                          json:"problem"`
	ObjectiveEvidence  string    `gorm:"column:objective_evidence;type:text"               json:"objectiveEvidence"`
	Location           string    `gorm:"column:location;type:varchar(255)"                 json:"location"`
	Reference          string    `gorm:"column:reference;type:text"                        json:"reference"`
	Process            string    `gorm:"column:process;type:varchar(50);default:'Outstanding'" json:"process"`
	Stage              int       `gorm:"column:stage;default:1"                            json:"stage"`
	// Workflow state disimpan sebagai JSON blob
	WorkflowData       string    `gorm:"column:workflow_data;type:text"                    json:"workflowData"`
	SubmittedAt        time.Time `gorm:"column:submitted_at;autoCreateTime"                json:"submittedAt"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime"                  json:"updatedAt"`
}

func (ANCRRequest) TableName() string {
	return "magang.ancr_requests"
}
