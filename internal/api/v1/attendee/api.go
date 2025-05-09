package attendee

import "github.com/eurofurence/reg-attendee-service/internal/api/v1/status"

type AttendeeDto struct {
	Id       uint   `json:"id"`       // badge number
	Nickname string `json:"nickname"` // fan name

	// name and address
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Street    string `json:"street"`
	Zip       string `json:"zip"`
	City      string `json:"city"`
	Country   string `json:"country"` // 2 letter ISO-3166-1 country code for the address (Alpha-2 code)
	State     string `json:"state"`

	// contact info
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Telegram string `json:"telegram"`
	Partner  string `json:"partner"`

	// personal data
	Birthday             string `json:"birthday"` // ISO date (format yyyy-MM-dd)
	Gender               string `json:"gender"`   // optional, one of male,female,other,notprovided
	Pronouns             string `json:"pronouns"` // optional
	TshirtSize           string `json:"tshirt_size"`
	SpokenLanguages      string `json:"spoken_languages"`      // configurable subset of configured language codes, comma separated (de,en)
	RegistrationLanguage string `json:"registration_language"` // one out of configurable subset of RFC 5646 locales (default en-US)

	// comma separated lists, allowed choices are convention dependent
	Flags        string         `json:"flags"`    // hc,anon,ev
	Packages     string         `json:"packages"` // room-none,attendance,stage,sponsor,sponsor2
	PackagesList []PackageState `json:"packages_list"`
	Options      string         `json:"options"` // art,anim,music,suit

	// comments
	UserComments string `json:"user_comments"`
}

type AttendeeMaxIdDto struct {
	MaxId uint `json:"max_id"`
}

type AttendeeIdList struct {
	Ids []uint `json:"ids"`
}

type DueDate struct {
	// The due date as an ISO date. Note that this is an accounting date, which does not have a time or a time zone.
	DueDate string `json:"due_date"`
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
	Ids                  []uint          `json:"ids,omitempty"`
	Nickname             string          `json:"nickname"`
	Name                 string          `json:"name"`
	Address              string          `json:"address"`
	Country              string          `json:"country"`
	BirthdayFrom         string          `json:"birthday_from"`
	BirthdayTo           string          `json:"birthday_to"`
	Email                string          `json:"email"`
	Telegram             string          `json:"telegram"`
	SpokenLanguages      map[string]int8 `json:"spoken_languages"`
	RegistrationLanguage map[string]int8 `json:"registration_language"`
	Flags                map[string]int8 `json:"flags"`
	Options              map[string]int8 `json:"options"`
	Packages             map[string]int8 `json:"packages"`
	UserComments         string          `json:"user_comments"`
	Status               []status.Status `json:"status"`
	Permissions          map[string]int8 `json:"permissions"`
	AdminComments        string          `json:"admin_comments"`
	AddInfo              map[string]int8 `json:"add_info"` // can only search for presence of a value for each area, Note: special area 'overdue'
	IdentitySubjects     []string        `json:"identity_subjects"`
}

// --- search result ---

type AttendeeSearchResultList struct {
	Attendees []AttendeeSearchResult `json:"attendees"`
}

type AttendeeSearchResult struct {
	Id                   uint           `json:"id"`
	BadgeId              *string        `json:"badge_id,omitempty"`
	Nickname             *string        `json:"nickname,omitempty"`
	FirstName            *string        `json:"first_name,omitempty"`
	LastName             *string        `json:"last_name,omitempty"`
	Street               *string        `json:"street,omitempty"`
	Zip                  *string        `json:"zip,omitempty"`
	City                 *string        `json:"city,omitempty"`
	Country              *string        `json:"country,omitempty"`
	State                *string        `json:"state,omitempty"`
	Email                *string        `json:"email,omitempty"`
	Phone                *string        `json:"phone,omitempty"`
	Telegram             *string        `json:"telegram,omitempty"`
	Partner              *string        `json:"partner,omitempty"`
	Birthday             *string        `json:"birthday,omitempty"`
	Gender               *string        `json:"gender,omitempty"`
	Pronouns             *string        `json:"pronouns,omitempty"`
	TshirtSize           *string        `json:"tshirt_size,omitempty"`
	SpokenLanguages      *string        `json:"spoken_languages,omitempty"`
	SpokenLanguagesList  []string       `json:"spoken_languages_list,omitempty"`
	RegistrationLanguage *string        `json:"registration_language,omitempty"`
	Flags                *string        `json:"flags,omitempty"`
	FlagsList            []string       `json:"flags_list,omitempty"`
	Options              *string        `json:"options,omitempty"`
	OptionsList          []string       `json:"options_list,omitempty"`
	Packages             *string        `json:"packages,omitempty"`
	PackagesList         []PackageState `json:"packages_list,omitempty"`
	UserComments         *string        `json:"user_comments,omitempty"`
	Status               *status.Status `json:"status,omitempty"`
	TotalDues            *int64         `json:"total_dues,omitempty"`
	PaymentBalance       *int64         `json:"payment_balance,omitempty"`
	CurrentDues          *int64         `json:"current_dues,omitempty"`
	DueDate              *string        `json:"due_date,omitempty"`
	Registered           *string        `json:"registered,omitempty"`
	AdminComments        *string        `json:"admin_comments,omitempty"`
	IdentitySubject      *string        `json:"identity_subject"`
	Avatar               *string        `json:"avatar"`
}

// --- flags/options/packages result ---

type ChoiceState struct {
	Present bool `json:"present"`
	Count   int  `json:"count"`
}

type PackageState struct {
	Name  string `json:"name"`
	Count int    `json:"count"` // defaults to 1 if unset in requests
}
