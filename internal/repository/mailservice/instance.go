package mailservice

var activeInstance MailService

func Create() {
	// TODO implement me with non-mock
	activeInstance = newMock()
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() MailService {
	return activeInstance
}
