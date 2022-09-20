package mailservice

var activeInstance MailService

func Create() {
	// TODO implement me with non-mock
	activeInstance = NewMock()
}

func CreateMock() Mock {
	instance := NewMock()
	activeInstance = instance
	return instance
}

func Get() MailService {
	return activeInstance
}
