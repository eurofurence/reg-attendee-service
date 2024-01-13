package config

type (
	DatabaseType string
	LogStyle     string
)

const (
	Inmemory DatabaseType = "inmemory"
	Mysql    DatabaseType = "mysql"

	Plain LogStyle = "plain"
	ECS   LogStyle = "ecs" // default
)

const StartTimeFormat = "2006-01-02T15:04:05-07:00"

const IsoDateFormat = "2006-01-02"

const HumanDateFormat = "02.01.2006"

type (
	// Application is the root configuration type
	Application struct {
		Service               ServiceConfig            `yaml:"service"`
		Server                ServerConfig             `yaml:"server"`
		Database              DatabaseConfig           `yaml:"database"`
		Security              SecurityConfig           `yaml:"security"`
		Logging               LoggingConfig            `yaml:"logging"`
		Choices               FlagsPkgOptConfig        `yaml:"choices"`
		AdditionalInfo        map[string]AddInfoConfig `yaml:"additional_info_areas"` // field name -> config
		TShirtSizes           []string                 `yaml:"tshirtsizes"`
		Birthday              BirthdayConfig           `yaml:"birthday"`
		GoLive                GoLiveConfig             `yaml:"go_live"`
		Dues                  DuesConfig               `yaml:"dues"`
		Countries             []string                 `yaml:"countries"`
		SpokenLanguages       []string                 `yaml:"spoken_languages"`
		RegistrationLanguages []string                 `yaml:"registration_languages"`
		Currency              string                   `yaml:"currency"`
		VatPercent            float64                  `yaml:"vat_percent"` // used for manual dues
	}

	// ServiceConfig contains configuration values
	// for service related tasks. E.g. URLs to downstream services
	ServiceConfig struct {
		Name            string `yaml:"name"`
		RegsysPublicUrl string `yaml:"regsys_public_url"` // used in emails
		PaymentService  string `yaml:"payment_service"`   // base url, usually http://localhost:nnnn, will use in-memory-mock if unset
		MailService     string `yaml:"mail_service"`      // base url, usually http://localhost:nnnn, will use in-memory-mock if unset
		AuthService     string `yaml:"auth_service"`      // base url, usually http://localhost:nnnn, will skip userinfo checks if unset
	}

	// ServerConfig contains all values for http configuration
	ServerConfig struct {
		Address      string `yaml:"address"`
		Port         string `yaml:"port"`
		ReadTimeout  int    `yaml:"read_timeout_seconds"`
		WriteTimeout int    `yaml:"write_timeout_seconds"`
		IdleTimeout  int    `yaml:"idle_timeout_seconds"`
	}

	// DatabaseConfig configures which db to use (mysql, inmemory)
	// and how to connect to it (needed for mysql only)
	DatabaseConfig struct {
		Use        DatabaseType `yaml:"use"`
		Username   string       `yaml:"username"`
		Password   string       `yaml:"password"`
		Database   string       `yaml:"database"`
		Parameters []string     `yaml:"parameters"`
	}

	// SecurityConfig configures everything related to security
	SecurityConfig struct {
		Fixed        FixedTokenConfig    `yaml:"fixed_token"`
		Oidc         OpenIdConnectConfig `yaml:"oidc"`
		Cors         CorsConfig          `yaml:"cors"`
		RequireLogin bool                `yaml:"require_login_for_reg"`
	}

	FixedTokenConfig struct {
		Api string `yaml:"api"` // shared-secret for server-to-server backend authentication
	}

	OpenIdConnectConfig struct {
		IdTokenCookieName     string   `yaml:"id_token_cookie_name"`     // optional, but must both be set, then tokens are read from cookies
		AccessTokenCookieName string   `yaml:"access_token_cookie_name"` // optional, but must both be set, then tokens are read from cookies
		TokenPublicKeysPEM    []string `yaml:"token_public_keys_PEM"`    // a list of public RSA keys in PEM format, see https://github.com/Jumpy-Squirrel/jwks2pem for obtaining PEM from openid keyset endpoint
		AdminGroup            string   `yaml:"admin_group"`              // the group claim that supplies regsys admin rights
		EarlyRegGroup         string   `yaml:"early_reg_group"`          // optional, the group claim that turns on early registration
		Audience              string   `yaml:"audience"`
		Issuer                string   `yaml:"issuer"`
	}

	CorsConfig struct {
		DisableCors bool   `yaml:"disable"`
		AllowOrigin string `yaml:"allow_origin"`
	}

	// LoggingConfig configures logging
	LoggingConfig struct {
		Style    LogStyle `yaml:"style"`
		Severity string   `yaml:"severity"`
	}

	// FlagsPkgOptConfig configures the choices available to the attendees
	//
	// flags are choices that have some impact on how the registration is treated
	// (guest, e.V. membership, staff, ...)
	//
	// packages are stuff that costs money, such as sponsorship or housing options
	//
	// options are personal preferences (interested in music, fursuiter, ...) that do not
	// affect how the registration is treated.
	FlagsPkgOptConfig struct {
		Flags    map[string]ChoiceConfig `yaml:"flags"`
		Packages map[string]ChoiceConfig `yaml:"packages"`
		Options  map[string]ChoiceConfig `yaml:"options"`
	}

	ChoiceConfig struct {
		Description   string   `yaml:"description"`
		Price         int64    `yaml:"price"`
		VatPercent    float64  `yaml:"vat_percent"`
		Default       bool     `yaml:"default"`                // if set to true, is added to flags by default. Not available for admin only flags!
		AdminOnly     bool     `yaml:"admin_only"`             // this flag is kept under the adminInfo structure, so it is not visible to users
		ReadOnly      bool     `yaml:"read_only"`              // this flag is kept under the normal flags, thus visible to end user, but only admin can change it
		VisibleFor    []string `yaml:"visible_for"`            // list of permissions which allow seeing the flag/option/package. Admin can always see everything, "self" can always see non-admin_only, but you can add it for admin_only fields. This field also controls who else can see the info based on their permissions admin field. Example: "self,sponsordesk"
		Group         string   `yaml:"group"`                  // set if attendee has this group during initial registration
		Mandatory     bool     `yaml:"at-least-one-mandatory"` // one of these MUST be chosen (no constraint if not set on any choices)
		Constraint    string   `yaml:"constraint"`
		ConstraintMsg string   `yaml:"constraint_msg"`
	}

	// AddInfoConfig configures access permissions to an additional info field
	AddInfoConfig struct {
		SelfRead    bool     `yaml:"self_read"`
		SelfWrite   bool     `yaml:"self_write"`
		Permissions []string `yaml:"permissions"` // name of permission (in admin info) to grant access
		// could later also add groups
	}

	// BirthdayConfig is used for validation of attendee supplied birthday
	//
	// use it to exclude nonsensical values, or to exclude participants under a minimum age
	BirthdayConfig struct {
		Earliest string `yaml:"earliest"`
		Latest   string `yaml:"latest"`
	}

	// GoLiveConfig configures the time at which registration becomes available
	GoLiveConfig struct {
		StartIsoDatetime         string `yaml:"start_iso_datetime"`
		EarlyRegStartIsoDatetime string `yaml:"early_reg_start_iso_datetime"` // optional, only useful if you also set early_reg_role
	}

	// DuesConfig configures the due date calculations
	DuesConfig struct {
		EarliestDueDate string `yaml:"earliest_due_date"`
		LatestDueDate   string `yaml:"latest_due_date"` // inclusive
		DueDays         int    `yaml:"due_days"`
	}
)
