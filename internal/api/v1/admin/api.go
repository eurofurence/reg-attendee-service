package admin

type AdminInfoDto struct {
	Id string `json:"id"` // badge number - informational only, never read

	// comma separated lists of admin-only flags, allowed choices are convention dependent
	Flags string `json:"flags"` // security, dealer, ...

	// comma separated list of permissions, currently read_all, admin, payments, bans, announcements
	Permissions string `json:"permissions"`

	// comments
	AdminComments string `json:"admin_comments"`
}
