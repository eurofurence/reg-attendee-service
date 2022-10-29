package admin

type AdminInfoDto struct {
	Id string `json:"id"` // badge number - informational only, never read

	// comma separated lists of admin-only flags, allowed choices are convention dependent
	Flags string `json:"flags"` // security, dealer, ...

	// comma separated list of permissions, currently read_all, admin, payments, bans, announcements
	Permissions string `json:"permissions"`

	// comments
	AdminComments string `json:"admin_comments"`

	// Offset to book on the due amount, may be negative. If negative, will offset the highest VAT rates first. If positive, will be added at highest available VAT rate.
	ManualDues int64 `json:"manual_dues"`

	// Description to use for the manual dues booking.
	ManualDuesDescription string `json:"manual_dues_description"`
}
