package models

// DEV HTML REPORT OPEN SIGNAL v1
// Unique network signals per recipient prevent refreshes inflating the metric.
import "encoding/json"

const DevFileOpenSignal = "File Open Signal"

func AddUniqueDevFileOpenSignal(campaignID int64, email string, details EventDetails) error {
	var n int
	err := db.Model(&Event{}).Where("campaign_id=? AND email=? AND message=?", campaignID, email, DevFileOpenSignal).Count(&n).Error
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	encoded, err := json.Marshal(details)
	if err != nil {
		return err
	}
	return AddEvent(&Event{Email: email, Message: DevFileOpenSignal, Details: string(encoded)}, campaignID)
}
