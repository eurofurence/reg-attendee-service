package paymentservice

var activeInstance PaymentService

func Create() {
	// TODO implement me with non-mock
	activeInstance = newMock()
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() PaymentService {
	return activeInstance
}
