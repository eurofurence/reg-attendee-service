package addinfo

type AdditionalInfoFullArea struct {
	Area string `json:"area"` // the area that was requested

	Values map[string]string `json:"values"` // values by attendee id
}
