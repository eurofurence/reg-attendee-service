package config

// TODO implement config loading from config file

func AllowedFlags() []string {
	return []string{"hc", "anon", "ev"}
}

func AllowedPackages() []string {
	return []string{"room-none", "attendance", "stage", "sponsor", "sponsor2"}
}

func AllowedOptions() []string {
	return []string{"art", "anim", "music", "suit"}
}

func AllowedTshirtSizes() []string {
	return []string{"XS", "S", "M", "L", "XL", "XXL", "XXXL", "XXXXL"}
}
