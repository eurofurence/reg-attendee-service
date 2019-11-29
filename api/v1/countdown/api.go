package countdown

type CountdownResultDto struct {
	CurrentTimeIsoDateTime string `json:"currentTime"`
	TargetTimeIsoDateTime  string `json:"targetTime"`
	CountdownSeconds       int64  `json:"countdown"`
}
