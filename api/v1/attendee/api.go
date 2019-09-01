package attendee

type AttendeeDto struct {
	Id       string `json:"id"`       // badge number
	Nickname string `json:"nickname"` // fan name

	// name and address
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Street       string `json:"street"`
	Zip          string `json:"zip"`
	City         string `json:"city"`
	Country      string `json:"country"`       // 2 letter ISO-3166-1 country code for the address (Alpha-2 code)
	CountryBadge string `json:"country_badge"` // Alpha-2 code for the country to be shown on the badge
	State        string `json:"state"`

	// contact info
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Telegram string `json:"telegram"`

	// personal data
	Birthday   string `json:"birthday"` // ISO date (format yyyy-MM-dd)
	Gender     string `json:"gender"`   // optional, one of male,female,other,notprovided
	TshirtSize string `json:"tshirt_size"`

	// comma separated lists, allowed choices are convention dependent
	Flags    string `json:"flags"`    // hc,anon,ev
	Packages string `json:"packages"` // room-none,attendance,stage,sponsor,sponsor2
	Options  string `json:"options"`  // art,anim,music,suit

	// comments
	UserComments string `json:"user_comments"`
}
