package models

type DocumentTemplate struct {
	ID          uint64 `gorm:"primaryKey" json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Format      string `json:"format"`
	Size        string `json:"size"`
	FileName    string `json:"fileName"`
	FileData    string `gorm:"type:text" json:"fileData"`
}

func (DocumentTemplate) TableName() string { return "magang.document_templates" }
