package notifications

type Notifier interface {
	DispatchNotification(event interface{}) error
}
