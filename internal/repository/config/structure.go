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
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

type loggingConfig struct {
	Severity string `yaml:"severity"`
}

type fixedTokenConfig struct {
	Admin      string `yaml:"admin"`
	User       string `yaml:"user"`
	InitialReg string `yaml:"reg"`
}

type securityConfig struct {
	Use         string           `yaml:"use"` // fixed-token, currently only supported value
	Fixed       fixedTokenConfig `yaml:"fixed"`
	DisableCors bool             `yaml:"disable_cors"`
}

type ChoiceConfig struct {
	Description   string  `yaml:"description"`
	HelpUrl       string  `yaml:"help_url"`
	PriceEarly    float64 `yaml:"price_early"`
	PriceLate     float64 `yaml:"price_late"`
	PriceAtCon    float64 `yaml:"price_atcon"`
	VatPercent    float64 `yaml:"vat_percent"`
	Default       bool    `yaml:"default"`
	AdminOnly     bool    `yaml:"admin_only"`
	ReadOnly      bool    `yaml:"read_only"` // but admin can still remove this
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
	StartIsoDatetime string `yaml:"start_iso_datetime"`
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
}
