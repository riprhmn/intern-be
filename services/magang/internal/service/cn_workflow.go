package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"magang-be/services/magang/internal/models"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var ErrCNForbidden = errors.New("Anda tidak berwenang mengakses atau memproses CN ini")
var ErrCNConflict = errors.New("Data CN berubah. Muat ulang sebelum melanjutkan")

const cnWorkflowMasterMapping = "MASTER_MAPPING"

func ValidateCNDocument(name, data string) error {
	ext := strings.ToLower(filepath.Ext(name))
	allowed := map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".png": true, ".jpg": true, ".jpeg": true}
	if !allowed[ext] || strings.TrimSpace(name) == "" {
		return errors.New("Dokumen wajib diunggah dalam format PDF, Word, Excel, PNG atau JPG")
	}
	if len(data) > 14*1024*1024 {
		return errors.New("Ukuran dokumen maksimal 10 MB")
	}
	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:") || !strings.HasSuffix(parts[0], ";base64") {
		return errors.New("Data dokumen tidak valid")
	}
	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(raw) == 0 || len(raw) > 10*1024*1024 {
		return errors.New("Data dokumen tidak valid atau melebihi 10 MB")
	}
	return nil
}
func CNCanRead(cn *models.ChangeNote, u *models.User) bool {
	if u.Role == "admin" || cn.UserID == u.ID || cn.IsPIC == u.ID || sameCNDepartment(cn.Department, u.Department) {
		return true
	}
	for _, st := range cn.Stages {
		if st.UserID == u.ID || st.DelegateID == u.ID {
			return true
		}
	}
	return false
}

func sameCNDepartment(left, right string) bool {
	return normalizeCNAccountField(left) != "" && normalizeCNAccountField(left) == normalizeCNAccountField(right)
}
func CNCurrentStage(cn *models.ChangeNote) int {
	for i, st := range cn.Stages {
		if st.Status != "APPROVED" {
			return i
		}
	}
	return -1
}
func CNCanDecide(cn *models.ChangeNote, u *models.User) bool {
	i := CNCurrentStage(cn)
	if !u.IsActive || cn.Status != "WAITING_APPROVAL" || i < 0 || cnNeedsReregistration(cn) {
		return false
	}
	if cn.Stages[i].Label == "Initiator (2nd)" {
		return cn.UserID == u.ID && cn.Stages[i].UserID == u.ID
	}
	return cn.Stages[i].UserID == u.ID || cn.Stages[i].DelegateID == u.ID
}
func (s *ChangeNoteService) Visible(u *models.User, search, mode, department string, page, limit int) ([]models.ChangeNote, int64, error) {
	if err := s.refreshPendingCNDelegations(); err != nil {
		return nil, 0, err
	}
	q := applyCNMode(s.visibleQueryForMode(u, mode), mode)
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("(initiator ILIKE ? OR department ILIKE ? OR change_pertains_to ILIKE ? OR document_code ILIKE ? OR registration_number ILIKE ?)", like, like, like, like, like)
	}
	if strings.TrimSpace(department) != "" {
		q = q.Where("LOWER(TRIM(department)) = ?", strings.ToLower(strings.TrimSpace(department)))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	notes := []models.ChangeNote{}
	err := q.Omit("document_data", "processed_document_data", "revised_documents", "related_documents").Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&notes).Error
	return notes, total, err
}

func applyCNMode(q *gorm.DB, mode string) *gorm.DB {
	switch mode {
	case "process":
		return q.Where("status IN ?", []string{"SUBMITTED", "WAITING_ISO", "WAITING_APPROVAL", "WAITING_PUBLISH"})
	case "iso_process":
		return q.Where("status IN ?", []string{"SUBMITTED", "WAITING_ISO", "REJECTED", "WAITING_APPROVAL", "WAITING_PUBLISH"})
	case "approval":
		return q.Where("status IN ?", []string{"WAITING_APPROVAL", "WAITING_PUBLISH"})
	case "history":
		return q.Where("status IN ?", []string{"APPROVED", "REJECTED"})
	case "active":
		return q.Where("status = ? AND document_code <> ''", "APPROVED").Where("NOT EXISTS (SELECT 1 FROM magang.change_notes newer WHERE newer.document_code = change_notes.document_code AND newer.status = 'APPROVED' AND (newer.updated_at, newer.id) > (change_notes.updated_at, change_notes.id))")
	default:
		return q
	}
}

func (s *ChangeNoteService) Departments(u *models.User, mode string) ([]string, error) {
	departments := []string{}
	err := applyCNMode(s.visibleQueryForMode(u, mode), mode).
		Where("TRIM(department) <> ''").
		Distinct().Order("department").Pluck("department", &departments).Error
	return departments, err
}

type CNAction struct {
	EffectedDate string `json:"effected_date"`

	Form                *CNForm `json:"form"`
	DocumentCode        string  `json:"document_code"`
	DelegateID          uint64  `json:"delegate_id"`
	Action              string  `json:"action"`
	Version             uint64  `json:"version"`
	Comment             string  `json:"comment"`
	DocumentName        string  `json:"document_name"`
	DocumentData        string  `json:"document_data"`
	IsPIC               uint64  `json:"is_pic"`
	FinalResultDocument string  `json:"final_result_document"`
}

type CNForm struct {
	TWA                 string              `json:"twa"`
	ProposedChange      string              `json:"proposed_change"`
	ReasonForChange     string              `json:"reason_for_change"`
	SupportingDocuments string              `json:"supporting_documents"`
	RevisedDocuments    []models.CNDocument `json:"revised_documents"`
	RelatedDocuments    []models.CNDocument `json:"related_documents"`
}

func validateCNForm(form *CNForm) error {
	if form == nil {
		return errors.New("Data form wajib diisi")
	}
	if len(form.TWA) > 150 || len(form.ProposedChange) > 20000 || len(form.ReasonForChange) > 20000 || len(form.SupportingDocuments) > 20000 {
		return errors.New("Isian form terlalu panjang")
	}
	if len(form.RevisedDocuments)+len(form.RelatedDocuments) > 20 {
		return errors.New("Maksimal 20 lampiran per CN")
	}
	total := 0
	for _, group := range [][]models.CNDocument{form.RevisedDocuments, form.RelatedDocuments} {
		for _, doc := range group {
			if err := ValidateCNDocument(doc.Name, doc.Data); err != nil {
				return err
			}
			total += len(doc.Data)
			if strings.TrimSpace(doc.DocumentToRevise) == "" {
				return errors.New("Document to Revise wajib diisi")
			}
		}
	}
	if total > 14*1024*1024 {
		return errors.New("Total ukuran lampiran maksimal 10 MB")
	}
	for _, doc := range form.RevisedDocuments {
		if strings.TrimSpace(doc.RevisionVersion) == "" {
			return errors.New("Revision Version wajib diisi")
		}
	}
	for _, doc := range form.RelatedDocuments {
		if strings.TrimSpace(doc.Department) == "" {
			return errors.New("Related Dept/Section wajib diisi")
		}
	}
	return nil
}

func (s *ChangeNoteService) Act(id uint64, u *models.User, a CNAction) (*models.ChangeNote, error) {
	if err := s.refreshPendingCNDelegations(); err != nil {
		return nil, err
	}
	var cn models.ChangeNote
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cn, id).Error; err != nil {
			return err
		}
		if !CNCanRead(&cn, u) {
			return ErrCNForbidden
		}
		if cn.Version != a.Version {
			return ErrCNConflict
		}
		switch a.Action {
		case "update_form":
			if u.Role != "admin" && cn.UserID != u.ID {
				return ErrCNForbidden
			}
			masterMigration := u.Role == "admin" && cnNeedsReregistration(&cn)
			if cn.Status != "SUBMITTED" && cn.Status != "WAITING_ISO" && cn.Status != "REJECTED" && !masterMigration {
				return errors.New("Form hanya dapat diperbarui saat menunggu proses ISO")
			}
			resumeApproval := cn.Status == "REJECTED" && validCNHierarchy(&cn)
			if err := validateCNUploadOwnership(&cn, u, a.Form); err != nil {
				return err
			}
			// Related Dept/Section selalu mengikuti organisasi initiator yang tersimpan di database.
			for i := range a.Form.RelatedDocuments {
				a.Form.RelatedDocuments[i].Department = cnRelatedTarget(&cn)
			}
			if err := validateCNForm(a.Form); err != nil {
				return err
			}
			cn.TWA = strings.TrimSpace(a.Form.TWA)
			cn.ProposedChange = strings.TrimSpace(a.Form.ProposedChange)
			cn.ReasonForChange = strings.TrimSpace(a.Form.ReasonForChange)
			cn.SupportingDocuments = strings.TrimSpace(a.Form.SupportingDocuments)
			cn.RevisedDocuments = a.Form.RevisedDocuments
			cn.RelatedDocuments = a.Form.RelatedDocuments
			cn.EffectedDate = ""
			cn.FinalResultDocument = ""
			if resumeApproval {
				restartCNApproval(&cn)
				cn.Status = "WAITING_APPROVAL"
				a.Comment = "Revisi disubmit; approval dimulai kembali tanpa registrasi ulang ISO"
			} else if u.Role == "admin" {
				if err := startCNAutomaticApproval(tx, &cn); err != nil {
					return err
				}
				a.Action = "process"
				a.Comment = "Document to Revise disimpan; hierarchy dibentuk otomatis oleh sistem"
			} else {
				cn.Status = "WAITING_ISO"
				cn.Stages = []models.ApprovalStage{}
				cn.IsPIC = 0
				a.Comment = "Form Change Note diperbarui"
			}
		case "process":
			if u.Role != "admin" {
				return ErrCNForbidden
			}
			if cn.Status != "SUBMITTED" && cn.Status != "WAITING_ISO" && cn.Status != "REJECTED" && !cnNeedsReregistration(&cn) {
				return errors.New("CN tidak sedang menunggu proses ISO")
			}
			if err := startCNAutomaticApproval(tx, &cn); err != nil {
				return err
			}
			a.Comment = "Hierarchy dibentuk otomatis oleh sistem"
		case "delegate":
			if u.Role != "admin" || cn.Status != "WAITING_APPROVAL" {
				return ErrCNForbidden
			}
			i := CNCurrentStage(&cn)
			if i < 0 {
				return errors.New("Tidak ada tahap menunggu")
			}
			if cn.Stages[i].Label == "Initiator (2nd)" {
				return errors.New("Tahap Initiator (2nd) wajib dilakukan pemohon dan tidak dapat didelegasikan")
			}
			if a.DelegateID == cn.Stages[i].UserID {
				return errors.New("Delegasi harus berbeda dari approver")
			}
			if a.DelegateID != 0 {
				var delegate models.User
				if err := tx.First(&delegate, a.DelegateID).Error; err != nil || !delegate.IsActive || delegate.Role != "approval" {
					return errors.New("Delegasi harus akun approval aktif")
				}
			}
			cn.Stages[i].DelegateID = a.DelegateID
			a.Comment = fmt.Sprintf("Delegasi tahap %s: user #%d", cn.Stages[i].Label, a.DelegateID)
		case "approve", "reject":
			if !CNCanDecide(&cn, u) {
				return ErrCNForbidden
			}
			if a.Action == "reject" && strings.TrimSpace(a.Comment) == "" {
				return errors.New("Alasan penolakan wajib diisi")
			}
			i := CNCurrentStage(&cn)
			if a.Action == "approve" && cn.Stages[i].Label == "Initiator (2nd)" {
				if err := validateEffectedDate(a.EffectedDate); err != nil {
					return err
				}
				cn.EffectedDate = a.EffectedDate
			}
			a.Comment = cn.Stages[i].Label + ": " + strings.TrimSpace(a.Comment)
			now := time.Now().UTC()
			cn.Stages[i].ActorID = u.ID
			cn.Stages[i].Comment = strings.TrimSpace(a.Comment)
			cn.Stages[i].ActedAt = &now
			if a.Action == "reject" {
				cn.Stages[i].Status = "REJECTED"
				cn.Status = "REJECTED"
			} else {
				cn.Stages[i].Status = "APPROVED"
				if CNCurrentStage(&cn) == -1 {
					cn.Status = "WAITING_PUBLISH"
				}
			}
		case "reject_final":
			if err := returnCNToInitiator(&cn, u, a.Comment); err != nil {
				return err
			}
		case "finalize":
			if err := validateCNPublication(&cn, u, a); err != nil {
				return err
			}
			cn.FinalResultDocument = strings.TrimSpace(a.FinalResultDocument)
			cn.Status = "APPROVED"
			a.Comment = "IS PIC publish final result: " + strings.TrimSpace(a.Comment)
		default:
			return errors.New("Tindakan CN tidak dikenal; status tidak dapat diubah langsung")
		}
		cn.Version++
		cn.History = append(cn.History, models.CNEvent{Action: a.Action, ActorID: u.ID, ActorName: u.FullName, Comment: strings.TrimSpace(a.Comment), At: time.Now().UTC()})
		return tx.Save(&cn).Error
	})
	return &cn, err
}

// Every active hierarchy must be a valid snapshot of Master Approving.
func cnNeedsReregistration(cn *models.ChangeNote) bool {
	return (cn.Status == "WAITING_APPROVAL" || cn.Status == "WAITING_PUBLISH") &&
		(cn.WorkflowSource != cnWorkflowMasterMapping || !validCNHierarchy(cn))
}

func restartCNApproval(cn *models.ChangeNote) {
	now := time.Now().UTC()
	for i := range cn.Stages {
		cn.Stages[i].Status = "WAITING_APPROVAL"
		cn.Stages[i].ActorID = 0
		cn.Stages[i].Comment = ""
		cn.Stages[i].ActedAt = nil
	}
	cn.Stages[0].Status = "APPROVED"
	cn.Stages[0].ActorID = cn.UserID
	cn.Stages[0].Comment = "Revision submitted"
	cn.Stages[0].ActedAt = &now
}
func validateEffectedDate(value string) error {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return errors.New("Effected date wajib diisi dengan tanggal valid")
	}
	return nil
}
func returnCNToInitiator(cn *models.ChangeNote, u *models.User, comment string) error {
	if !u.IsActive || cn.IsPIC == 0 || cn.IsPIC != u.ID || cn.Status != "WAITING_PUBLISH" {
		return ErrCNForbidden
	}
	if !validCNHierarchy(cn) || CNCurrentStage(cn) != -1 || cn.Stages[len(cn.Stages)-1].Label != "Initiator (2nd)" || cn.Stages[len(cn.Stages)-1].UserID != cn.UserID {
		return errors.New("Hierarchy belum siap untuk koreksi effected date")
	}
	if strings.TrimSpace(comment) == "" {
		return errors.New("Alasan koreksi effected date wajib diisi")
	}
	cn.Stages[len(cn.Stages)-1] = models.ApprovalStage{Label: "Initiator (2nd)", UserID: cn.UserID, Status: "WAITING_APPROVAL"}
	cn.Status = "WAITING_APPROVAL"
	return nil
}
func validateCNPublication(cn *models.ChangeNote, u *models.User, a CNAction) error {
	if !u.IsActive || cn.IsPIC == 0 || cn.IsPIC != u.ID || cn.Status != "WAITING_PUBLISH" {
		return ErrCNForbidden
	}
	if !validCNHierarchy(cn) || CNCurrentStage(cn) != -1 {
		return errors.New("Seluruh tahap approval wajib selesai")
	}
	if err := validateEffectedDate(cn.EffectedDate); err != nil {
		return err
	}
	if strings.TrimSpace(a.FinalResultDocument) == "" {
		return errors.New("Final result wajib diisi")
	}
	return nil
}
func startCNAutomaticApproval(tx *gorm.DB, cn *models.ChangeNote) error {
	if strings.TrimSpace(cn.ProposedChange) == "" {
		return errors.New("Pemohon wajib UPDATE DATA sebelum Tim ISO mengunggah Document to Revise")
	}
	if len(cn.RevisedDocuments) == 0 {
		return errors.New("ISO/admin wajib mengunggah Document to Revise")
	}
	document := cn.RevisedDocuments[0]
	if err := ValidateCNDocument(document.Name, document.Data); err != nil {
		return err
	}
	if strings.TrimSpace(document.DocumentToRevise) == "" {
		return errors.New("Document to Revise wajib diisi")
	}
	usedMaster, err := startCNMasterApproval(tx, cn)
	if err != nil {
		return err
	}
	if !usedMaster {
		section := strings.ToUpper(strings.TrimSpace(cn.Section))
		if section == "" {
			section = "(SECTION BELUM DIISI)"
		}
		return fmt.Errorf("Master Approving CN untuk Section Code '%s' belum tersedia", section)
	}
	return nil
}

func startCNMasterApproval(tx *gorm.DB, cn *models.ChangeNote) (bool, error) {
	section := strings.ToUpper(strings.TrimSpace(cn.Section))
	if section == "" {
		return false, nil
	}
	var mappings []models.ApprovalMapping
	if err := tx.Preload("User").Where(
		"approval_type = ? AND section_code = ? AND is_active = ?", "CN", section, true,
	).Order("sequence, id").Find(&mappings).Error; err != nil {
		return false, err
	}
	if len(mappings) == 0 {
		return false, nil
	}
	for i := range mappings {
		if mappings[i].Sequence != i+1 {
			return false, errors.New("Master mapping CN harus berurutan mulai dari SEQ 1")
		}
		user := mappings[i].User
		if !user.IsActive || user.Role != "approval" || strings.TrimSpace(user.CodeName) == "" {
			return false, fmt.Errorf("Approver master mapping SEQ %d tidak aktif atau tidak valid", mappings[i].Sequence)
		}
	}

	var isoMappings []models.ApprovalMapping
	if err := tx.Preload("User").Where(
		"approval_type = ? AND section_code = ? AND is_active = ?", "CN", "ISO", true,
	).Order("sequence, id").Find(&isoMappings).Error; err != nil {
		return false, err
	}
	if len(isoMappings) != 3 {
		return false, errors.New("Konfigurasi ISO belum divalidasi. Wajib buat Master Mapping untuk Section Code 'ISO' dengan urutan SEQ 1 (IS Section), SEQ 2 (MQD), SEQ 3 (TDIV)")
	}
	for i := range isoMappings {
		if isoMappings[i].Sequence != i+1 {
			return false, errors.New("Master mapping ISO harus berurutan SEQ 1, 2, 3")
		}
		if !isoMappings[i].User.IsActive || isoMappings[i].User.Role != "approval" {
			return false, fmt.Errorf("Approver ISO SEQ %d tidak valid", isoMappings[i].Sequence)
		}
	}

	at := cn.CreatedAt
	for _, event := range cn.History {
		if event.ActorID == cn.UserID && (event.Action == "submit" || event.Action == "update_form") {
			at = event.At
		}
	}
	stages := []models.ApprovalStage{{Label: "Initiator", UserID: cn.UserID, Status: "APPROVED", ActorID: cn.UserID, ActedAt: &at, Comment: "Submit request"}}
	isoUserIDs := make(map[uint64]bool)
	for _, iso := range isoMappings {
		isoUserIDs[iso.UserID] = true
	}

	for _, mapping := range mappings {
		if isoUserIDs[mapping.UserID] {
			continue
		}
		delegateID, err := activeCNDelegate(tx, mapping.UserID, section)
		if err != nil {
			return false, err
		}
		title := strings.TrimSpace(mapping.User.Title)
		if title == "" {
			title = strings.TrimSpace(mapping.User.CodeName)
		}
		stages = append(stages, models.ApprovalStage{
			Label:  title,
			UserID: mapping.UserID, DelegateID: delegateID, Status: "WAITING_APPROVAL",
		})
	}
	stages = append(stages, models.ApprovalStage{Label: "IS Section Head", UserID: isoMappings[0].UserID, Status: "WAITING_APPROVAL"})
	stages = append(stages, models.ApprovalStage{Label: "MQD Manager", UserID: isoMappings[1].UserID, Status: "WAITING_APPROVAL"})
	stages = append(stages, models.ApprovalStage{Label: "TDIV Head", UserID: isoMappings[2].UserID, Status: "WAITING_APPROVAL"})
	stages = append(stages, models.ApprovalStage{Label: "Initiator (2nd)", UserID: cn.UserID, Status: "WAITING_APPROVAL"})
	cn.Stages = stages
	cn.IsPIC = isoMappings[0].UserID
	if strings.TrimSpace(cn.DocumentCode) == "" {
		year := cnYear(cn.CreatedAt)
		number, err := nextCNSerial(tx, cnNumberKindChangeNote, year)
		if err != nil {
			return false, err
		}
		cn.DocumentCode = formatChangeNoteNumber(year, number)
	}
	cn.ProcessDescription = "Hierarchy disalin dari master mapping CN · " + section
	cn.ProcessedDocumentName = cn.RevisedDocuments[0].Name
	cn.ProcessedDocumentData = cn.RevisedDocuments[0].Data
	cn.EffectedDate = ""
	cn.FinalResultDocument = ""
	cn.WorkflowSource = cnWorkflowMasterMapping
	cn.WorkflowSection = section
	cn.Status = "WAITING_APPROVAL"
	return true, nil
}

func activeCNDelegate(tx *gorm.DB, fromUserID uint64, section string) (uint64, error) {
	var delegation models.ApprovalDelegation
	now := time.Now().UTC()
	secNorm := strings.TrimSpace(section)
	err := tx.Where(
		"from_user_id = ? AND is_active = ? AND approval_type IN ? AND (LOWER(TRIM(section_code)) = LOWER(?) OR UPPER(TRIM(section_code)) = 'ALL') AND (starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at >= ?)",
		fromUserID, true, []string{"CN", "ALL"}, secNorm, now, now,
	).Order(clause.Expr{SQL: "CASE WHEN approval_type = 'CN' THEN 0 ELSE 1 END, CASE WHEN LOWER(TRIM(section_code)) = LOWER(?) THEN 0 WHEN section_code = 'ALL' THEN 1 ELSE 2 END, created_at DESC", Vars: []interface{}{secNorm}, WithoutParentheses: true}).
		First(&delegation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return delegation.ToUserID, nil
}
func normalizeCNAccountField(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func cnRelatedTarget(cn *models.ChangeNote) string {
	department := strings.TrimSpace(cn.Department)
	section := strings.TrimSpace(cn.Section)
	if section == "" {
		return department
	}
	return department + " / " + section
}

// refreshPendingCNDelegations keeps delegation access dynamic while the owner
// and sequence remain the immutable transaction snapshot. Completed stages are
// never rewritten, preserving who was delegated when the decision happened.
func (s *ChangeNoteService) refreshPendingCNDelegations() error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		var notes []models.ChangeNote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
			"status = ? AND workflow_source = ?", "WAITING_APPROVAL", cnWorkflowMasterMapping,
		).Find(&notes).Error; err != nil {
			return err
		}
		for i := range notes {
			changed := false
			sec := notes[i].WorkflowSection
			for stageIndex := 1; stageIndex < len(notes[i].Stages)-1; stageIndex++ {
				stage := &notes[i].Stages[stageIndex]
				if stage.Status != "WAITING_APPROVAL" || stage.ActedAt != nil {
					continue
				}
				delegateID, err := activeCNDelegate(tx, stage.UserID, sec)
				if err != nil {
					return err
				}
				if stage.DelegateID != delegateID {
					stage.DelegateID = delegateID
					changed = true
				}
			}
			if changed {
				if err := tx.Model(&notes[i]).Select("Stages").Updates(&notes[i]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *ChangeNoteService) visibleQuery(u *models.User) *gorm.DB {
	q := s.repo.DB.Model(&models.ChangeNote{})
	switch u.Role {
	case "admin":
	case "approval":
		q = q.Where("is_pic = ? OR user_id = ? OR EXISTS (SELECT 1 FROM jsonb_array_elements(COALESCE(NULLIF(stages, ''), '[]')::jsonb) st WHERE (st->>'user_id')::bigint = ? OR (st->>'delegate_id')::bigint = ?)", u.ID, u.ID, u.ID, u.ID)
	case "external_audit":
		q = q.Where("status = ? OR user_id = ? OR is_pic = ?", "APPROVED", u.ID, u.ID)
	default:
		q = q.Where("user_id = ? OR is_pic = ?", u.ID, u.ID)
	}
	return q
}

func (s *ChangeNoteService) visibleQueryForMode(u *models.User, mode string) *gorm.DB {
	q := s.repo.DB.Model(&models.ChangeNote{})
	if mode == "approval" {
		if u.Role == "admin" {
			return q
		}
		return cnApprovalTaskQuery(q, u.ID, false)
	}
	if cnModeUsesDepartmentScope(mode, u.Role) {
		department := strings.ToLower(strings.TrimSpace(u.Department))
		if mode == "process" {
			return q.Where(`LOWER(TRIM(department)) = ? OR user_id = ? OR is_pic = ? OR EXISTS (
				SELECT 1 FROM jsonb_array_elements(COALESCE(NULLIF(stages, ''), '[]')::jsonb) stage
				WHERE COALESCE(NULLIF(stage->>'user_id', ''), '0')::bigint = ?
					OR COALESCE(NULLIF(stage->>'delegate_id', ''), '0')::bigint = ?
			)`, department, u.ID, u.ID, u.ID, u.ID)
		}
		if department == "" {
			return q.Where("1 = 0")
		}
		return q.Where("LOWER(TRIM(department)) = ?", department)
	}
	return s.visibleQuery(u)
}

func cnApprovalTaskQuery(q *gorm.DB, userID uint64, exclude bool) *gorm.DB {
	condition := `((status = 'WAITING_APPROVAL' AND (
		COALESCE((
			SELECT COALESCE(NULLIF(current_stage.stage->>'user_id', ''), '0')::bigint = ?
				OR COALESCE(NULLIF(current_stage.stage->>'delegate_id', ''), '0')::bigint = ?
			FROM jsonb_array_elements(COALESCE(NULLIF(stages, ''), '[]')::jsonb)
				WITH ORDINALITY AS current_stage(stage, position)
			WHERE COALESCE(current_stage.stage->>'status', '') <> 'APPROVED'
			ORDER BY current_stage.position
			LIMIT 1
		), false)
	)) OR (status = 'WAITING_PUBLISH' AND is_pic = ?))`
	if exclude {
		return q.Where("NOT "+condition, userID, userID, userID)
	}
	return q.Where(condition, userID, userID, userID)
}

func cnModeUsesDepartmentScope(mode, role string) bool {
	return (mode == "dashboard" || mode == "process" || mode == "report" || mode == "active") && role != "admin"
}

func (s *ChangeNoteService) Summary(u *models.User) (map[string]int64, error) {
	var rows []struct {
		Status string
		Total  int64
	}
	if err := s.visibleQuery(u).Select("status, count(*) as total").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	_, activeTotal, err := s.Visible(u, "", "active", "", 1, 1)
	if err != nil {
		return nil, err
	}
	result := map[string]int64{"active_documents": activeTotal, "total": 0, "WAITING_ISO": 0, "WAITING_APPROVAL": 0, "WAITING_PUBLISH": 0, "APPROVED": 0, "REJECTED": 0}
	for _, row := range rows {
		key := row.Status
		if key == "SUBMITTED" {
			key = "WAITING_ISO"
		}
		result[key] += row.Total
		result["total"] += row.Total
	}
	return result, nil
}

func (s *ChangeNoteService) Templates() ([]models.DocumentTemplate, error) {
	rows := []models.DocumentTemplate{}
	err := s.repo.DB.Omit("file_data").Order("id DESC").Find(&rows).Error
	return rows, err
}
func (s *ChangeNoteService) Template(id uint64) (*models.DocumentTemplate, error) {
	var row models.DocumentTemplate
	err := s.repo.DB.First(&row, id).Error
	return &row, err
}
func (s *ChangeNoteService) SaveTemplate(row *models.DocumentTemplate) error {
	if row.ID != 0 {
		if _, err := s.Template(row.ID); err != nil {
			return err
		}
	}
	return s.repo.DB.Save(row).Error
}
func (s *ChangeNoteService) DeleteTemplate(id uint64) error {
	return s.repo.DB.Delete(&models.DocumentTemplate{}, id).Error
}

func validateCNUploadOwnership(cn *models.ChangeNote, u *models.User, form *CNForm) error {
	if form == nil {
		return errors.New("Data form wajib diisi")
	}
	equal := func(a, b []models.CNDocument) bool { return len(a) == 0 && len(b) == 0 || reflect.DeepEqual(a, b) }
	// An ISO/admin who submitted the CN is still its initiator and may update
	// the request fields. The ISO-only restriction applies when processing a CN
	// submitted by somebody else.
	if u.Role == "admin" && u.ID != cn.UserID && (strings.TrimSpace(cn.TWA) != strings.TrimSpace(form.TWA) || strings.TrimSpace(cn.ProposedChange) != strings.TrimSpace(form.ProposedChange) || strings.TrimSpace(cn.ReasonForChange) != strings.TrimSpace(form.ReasonForChange) || strings.TrimSpace(cn.SupportingDocuments) != strings.TrimSpace(form.SupportingDocuments)) {
		return errors.New("Tim ISO/admin hanya dapat mengubah Document to Revise")
	}
	if u.Role != "admin" && !equal(cn.RevisedDocuments, form.RevisedDocuments) {
		return errors.New("Document to Revise hanya dapat diunggah atau diubah ISO/admin")
	}
	if u.ID != cn.UserID && !equal(cn.RelatedDocuments, form.RelatedDocuments) {
		return errors.New("Relevant Document to Revise hanya dapat diunggah atau diubah pengaju")
	}
	return nil
}

func validCNHierarchy(cn *models.ChangeNote) bool {
	n := len(cn.Stages)
	if cn.WorkflowSource != cnWorkflowMasterMapping || n < 3 || cn.Stages[0].Label != "Initiator" || cn.Stages[n-1].Label != "Initiator (2nd)" || cn.Stages[n-1].UserID != cn.UserID {
		return false
	}
	for i := 1; i < n-1; i++ {
		if cn.Stages[i].UserID == 0 {
			return false
		}
	}
	return true
}
