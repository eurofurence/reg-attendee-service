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
	Partner  string `json:"partner"`

	// personal data
	Birthday   string `json:"birthday"` // ISO date (format yyyy-MM-dd)
	Gender     string `json:"gender"`   // optional, one of male,female,other,notprovided
	Pronouns   string `json:"pronouns"` // optional
	TshirtSize string `json:"tshirt_size"`

	// comma separated lists, allowed choices are convention dependent
	Flags    string `json:"flags"`    // hc,anon,ev
	Packages string `json:"packages"` // room-none,attendance,stage,sponsor,sponsor2
	Options  string `json:"options"`  // art,anim,music,suit

	// comments
	UserComments string `json:"user_comments"`
}

type AttendeeMaxIdDto struct {
	MaxId uint `json:"max_id"`
}

type AttendeeIdList struct {
	Ids []int64 `json:"ids"`
}

// --- search criteria ---

type AttendeeSearchCriteria struct {
	MatchAny   []AttendeeSearchSingleCriterion `json:"match_any"`
	MinId      uint                            `json:"min_id"`
	MaxId      uint                            `json:"max_id"`
	NumResults uint                            `json:"num_results"`
	FillFields []string                        `json:"fill_fields"`
	SortBy     string                          `json:"sort_by"`
	SortOrder  string                          `json:"sort_order"`
}

type AttendeeSearchSingleCriterion struct {
	Ids          []uint          `json:"ids,omitempty"`
	Nickname     string          `json:"nickname"`
	Name         string          `json:"name"`
	Address      string          `json:"address"`
	Country      string          `json:"country"`
	CountryBadge string          `json:"country_badge"`
	Email        string          `json:"email"`
	Telegram     string          `json:"telegram"`
	Flags        map[string]int8 `json:"flags"`
	Options      map[string]int8 `json:"options"`
	Packages     map[string]int8 `json:"packages"`
	UserComments string          `json:"user_comments"`
}

// --- search result ---

type AttendeeSearchResultList struct {
	Attendees []AttendeeSearchResult `json:"attendees"`
}

type AttendeeSearchResult struct {
	Id             int64   `json:"id"`
	BadgeId        *string `json:"badge_id,omitempty"`
	Nickname       *string `json:"nickname,omitempty"`
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	Street         *string `json:"street,omitempty"`
	Zip            *string `json:"zip,omitempty"`
	City           *string `json:"city,omitempty"`
	Country        *string `json:"country,omitempty"`
	CountryBadge   *string `json:"country_badge,omitempty"`
	State          *string `json:"state,omitempty"`
	Email          *string `json:"email,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Telegram       *string `json:"telegram,omitempty"`
	Partner        *string `json:"partner,omitempty"`
	Birthday       *string `json:"birthday,omitempty"`
	Gender         *string `json:"gender,omitempty"`
	Pronouns       *string `json:"pronouns,omitempty"`
	TshirtSize     *string `json:"tshirt_size,omitempty"`
	Flags          *string `json:"flags,omitempty"`
	Options        *string `json:"options,omitempty"`
	Packages       *string `json:"packages,omitempty"`
	UserComments   *string `json:"user_comments,omitempty"`
	Status         *string `json:"status,omitempty"`
	TotalDues      *int64  `json:"total_dues,omitempty"`
	PaymentBalance *int64  `json:"payment_balance,omitempty"`
	CurrentDues    *int64  `json:"current_dues,omitempty"`
	DueDate        *string `json:"due_date,omitempty"`
}
