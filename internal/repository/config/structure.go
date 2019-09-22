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

type choiceConfig struct {
	Description string  `yaml:"description"`
	HelpUrl     string  `yaml:"help_url"`
	PriceEarly  float64 `yaml:"price_early"`
	PriceLate   float64 `yaml:"price_late"`
	PriceAtCon  float64 `yaml:"price_atcon"`
	VatPercent  float64 `yaml:"vat_percent"`
	Default     bool    `yaml:"default"`
	AdminOnly   bool    `yaml:"admin_only"`
	ReadOnly    bool    `yaml:"read_only"`
	Constraint  string  `yaml:"constraint"`
}

type flagsPkgOptConfig struct {
	Flags    map[string]choiceConfig `yaml:"flags"`
	Packages map[string]choiceConfig `yaml:"packages"`
	Options  map[string]choiceConfig `yaml:"options"`
}

type conf struct {
	Database databaseConfig    `yaml:"database"`
	Server   serverConfig      `yaml:"server"`
	Choices  flagsPkgOptConfig `yaml:"choices"`
	Logging  loggingConfig     `yaml:"logging"`
}
