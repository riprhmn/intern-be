package service

import (
	"testing"

	"magang-be/services/magang/internal/models"
)

func TestValidateCNUploadOwnershipAllowsISOInitiatorToUpdateRequest(t *testing.T) {
	cn := &models.ChangeNote{
		UserID:              7,
		TWA:                 "General",
		ProposedChange:      "Usulan lama",
		ReasonForChange:     "Alasan lama",
		SupportingDocuments: "Dokumen lama",
	}
	user := &models.User{ID: 7, Role: "admin"}
	form := &CNForm{
		TWA:                 "Quality (QMS)",
		ProposedChange:      "Usulan baru",
		ReasonForChange:     "Alasan baru",
		SupportingDocuments: "Dokumen baru",
	}

	if err := validateCNUploadOwnership(cn, user, form); err != nil {
		t.Fatalf("ISO yang menjadi initiator seharusnya boleh memperbarui pengajuan: %v", err)
	}
}

func TestValidateCNUploadOwnershipRestrictsISOProcessingAnotherUsersCN(t *testing.T) {
	cn := &models.ChangeNote{UserID: 7, ProposedChange: "Usulan pemohon"}
	user := &models.User{ID: 9, Role: "admin"}
	form := &CNForm{ProposedChange: "Diubah ISO"}

	err := validateCNUploadOwnership(cn, user, form)
	if err == nil || err.Error() != "Tim ISO/admin hanya dapat mengubah Document to Revise" {
		t.Fatalf("ISO non-initiator seharusnya ditolak saat mengubah pengajuan, mendapat: %v", err)
	}
}
