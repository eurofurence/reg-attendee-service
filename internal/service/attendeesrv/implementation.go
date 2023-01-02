package attendeesrv

import "time"

type AttendeeServiceImplData struct {
	Now func() time.Time
}

var _ AttendeeService = (*AttendeeServiceImplData)(nil)

func New() AttendeeService {
	return &AttendeeServiceImplData{
		Now: time.Now,
	}
}
