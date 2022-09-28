package paymentservice

var activeInstance PaymentService

func Create() (err error) {
	activeInstance, err = newClient()
	return err
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() PaymentService {
	return activeInstance
}
