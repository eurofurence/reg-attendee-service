package config

type mysqlConfig struct {
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	Database   string   `yaml:"database"`
	Parameters []string `yaml:"parameters"`
}

type databaseConfig struct {
	Use   string      `yaml:"use"` // mysql or inmemory
	Mysql mysqlConfig `yaml:"mysql"`
}

type serverConfig struct {
	Address      string `yaml:"address"`
	Port         string `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout_seconds"`
	WriteTimeout int    `yaml:"write_timeout_seconds"`
	IdleTimeout  int    `yaml:"idle_timeout_seconds"`
}

type downstreamConfig struct {
	PaymentService string `yaml:"payment_service"` // base url, usually http://localhost:nnnn, will use in-memory-mock if unset
	MailService    string `yaml:"mail_service"`    // base url, usually http://localhost:nnnn, will use in-memory-mock if unset
}

type loggingConfig struct {
	Severity string `yaml:"severity"`
}

type fixedTokenConfig struct {
	Api string `yaml:"api"` // shared-secret for server-to-server backend authentication
}

type openIdConnectConfig struct {
	TokenCookieName    string   `yaml:"token_cookie_name"`     // optional, if set, the jwt token is also read from this cookie (useful for mixed web application setups, see reg-auth-service)
	TokenPublicKeysPEM []string `yaml:"token_public_keys_PEM"` // a list of public RSA keys in PEM format, see https://github.com/Jumpy-Squirrel/jwks2pem for obtaining PEM from openid keyset endpoint
	AdminRole          string   `yaml:"admin_role"`            // the role/group claim that supplies admin rights
	EarlyReg           string   `yaml:"early_reg_role"`        // optional, the role/group claim that turns on early staff registration
}

type securityConfig struct {
	Fixed        fixedTokenConfig    `yaml:"fixed_token"`
	Oidc         openIdConnectConfig `yaml:"oidc"`
	DisableCors  bool                `yaml:"disable_cors"`
	RequireLogin bool                `yaml:"require_login_for_reg"`
}

type ChoiceConfig struct {
	Description   string  `yaml:"description"`
	HelpUrl       string  `yaml:"help_url"`
	PriceEarly    float64 `yaml:"price_early"`
	PriceLate     float64 `yaml:"price_late"`
	PriceAtCon    float64 `yaml:"price_atcon"`
	VatPercent    float64 `yaml:"vat_percent"`
	Default       bool    `yaml:"default"`    // if set to true, is added to flags by default. Not available for admin only flags!
	AdminOnly     bool    `yaml:"admin_only"` // this flag is kept under the adminInfo structure, so it is not visible to users
	ReadOnly      bool    `yaml:"read_only"`  // this flag is kept under the normal flags, thus visible to end user, but only admin can change it
	Constraint    string  `yaml:"constraint"`
	ConstraintMsg string  `yaml:"constraint_msg"`
}

type flagsPkgOptConfig struct {
	Flags    map[string]ChoiceConfig `yaml:"flags"`
	Packages map[string]ChoiceConfig `yaml:"packages"`
	Options  map[string]ChoiceConfig `yaml:"options"`
}

type birthdayConfig struct {
	Earliest string `yaml:"earliest"`
	Latest   string `yaml:"latest"`
}

const StartTimeFormat = "2006-01-02T15:04:05-07:00"

type goLiveConfig struct {
	StartIsoDatetime         string `yaml:"start_iso_datetime"`
	EarlyRegStartIsoDatetime string `yaml:"early_reg_start_iso_datetime"` // optional, only useful if you also set early_reg_role
}

type conf struct {
	Database    databaseConfig    `yaml:"database"`
	Server      serverConfig      `yaml:"server"`
	Choices     flagsPkgOptConfig `yaml:"choices"`
	Logging     loggingConfig     `yaml:"logging"`
	Security    securityConfig    `yaml:"security"`
	TShirtSizes []string          `yaml:"tshirtsizes"`
	Birthday    birthdayConfig    `yaml:"birthday"`
	GoLive      goLiveConfig      `yaml:"go_live"`
	Countries   []string          `yaml:"countries"`
	Downstream  downstreamConfig  `yaml:"downstream"`
}
