package bans

type BanRule struct {
	Reason          string `json:"reason"`
	NamePattern     string `json:"name_pattern"`
	NicknamePattern string `json:"nickname_pattern"`
	EmailPattern    string `json:"email_pattern"`
}

type BanRuleList struct {
	Bans []BanRule `json:"bans"`
}
