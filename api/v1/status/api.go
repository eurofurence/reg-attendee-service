package status

type StatusDto struct {
	Status string `json:"status"` // new / approved / partially paid / paid / checked in / cancelled
}

type StatusHistoryDto struct {
	Id string `json:"id"` // badge number - informational only, never read

	// status history
	StatusHistory []StatusChange `json:"status_history"`
}

type StatusChange struct {
	Timestamp string `json:"timestamp"` // also gives registration date, and allows due date calculation
	Status    string `json:"status"`    // new / approved / partially paid / paid / checked in / cancelled
	Comment   string `json:"comment"`   // e.g. cancel reason
}
