package counts

type PackageCount struct {
	Pending   int `json:"pending"`
	Attending int `json:"attending"`
	Limit     int `json:"limit"`
}
