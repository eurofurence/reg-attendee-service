package status

type Status string

var (
	New           Status = "new"
	Approved      Status = "approved"
	PartiallyPaid Status = "partially paid"
	Paid          Status = "paid"
	CheckedIn     Status = "checked in"
	Cancelled     Status = "cancelled"
	Waiting       Status = "waiting"
	Deleted       Status = "deleted"
)

type StatusDto struct {
	Status Status `json:"status"`
}

type StatusHistoryDto struct {
	Id uint `json:"id"` // badge number - informational only, never read

	// status history
	StatusHistory []StatusChangeDto `json:"status_history"`
}

type StatusChangeDto struct {
	Timestamp string `json:"timestamp"` // also gives registration date, and allows due date calculation
	Status    Status `json:"status"`
	Comment   string `json:"comment"` // e.g. cancel reason
}
