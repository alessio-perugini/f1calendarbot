package domain

type SubscriptionService interface {
	Subscribe(int64)
	Unsubscribe(int64)
	GetAllSubscribedChats() []int64
}
