package paymentservice

var activeInstance PaymentService

func Create() {
	// TODO implement me with non-mock
	activeInstance = NewMock()
}

func CreateMock() Mock {
	instance := NewMock()
	activeInstance = instance
	return instance
}

func Get() PaymentService {
	return activeInstance
}
