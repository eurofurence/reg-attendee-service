package mailservice

var activeInstance MailService

func Create() (err error) {
	activeInstance, err = newClient()
	return err
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() MailService {
	return activeInstance
}
