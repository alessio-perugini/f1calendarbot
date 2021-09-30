package domain

type SubscriptionService interface {
	Subscribe(int)
	Unsubscribe(int)
	GetAllSubscribed() []int
}
