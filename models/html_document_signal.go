package models

// HTMLDocumentOpenSignal is a network request made by a locally opened
// downloaded HTML document. It is not proof a person read the document.
const HTMLDocumentOpenSignal = "HTML Open Signal"

// AddUniqueHTMLDocumentOpenSignal records a maximum of one signal per
// recipient and campaign. Reopening the file does not inflate the metric.
func AddUniqueHTMLDocumentOpenSignal(campaignID int64, email string) error {
	var count int
	err := db.Model(&Event{}).Where(
		"campaign_id=? AND email=? AND message=?",
		campaignID, email, HTMLDocumentOpenSignal,
	).Count(&count).Error
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	return AddEvent(&Event{
		Email:   email,
		Message: HTMLDocumentOpenSignal,
		Details: `{"meaning":"remote_image_request","evidence":"not_proof_of_local_open"}`,
	}, campaignID)
}
