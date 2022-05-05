package admin

import "net/url"

type AdminInfoDto struct {
	Id string `json:"id"` // badge number - informational only, never read

	// comma separated lists of admin-only flags, allowed choices are convention dependent
	Flags string `json:"flags"` // security, dealer, ...

	// status history
	StatusHistory []StatusChange `json:"status_history"`

	// comma separated list of permissions, currently read_all, admin, payments, bans, announcements
	Permissions string `json:"permissions"`

	// comments
	AdminComments string `json:"admin_comments"`
}

type StatusChange struct {
	Timestamp string `json:"timestamp"` // also gives registration date, and allows due date calculation
	Status    string `json:"status"`    // new / approved / partially paid / paid / checked in / cancelled
	Comment   string `json:"comment"`   // e.g. cancel reason
}

type ErrorDto struct {
	Timestamp string     `json:"timestamp"`
	RequestId string     `json:"requestid"`
	Message   string     `json:"message"`
	Details   url.Values `json:"details"`
}
