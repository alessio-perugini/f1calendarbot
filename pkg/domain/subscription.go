package domain

type SubscriptionService interface {
	Subscribe(interface{})
	Unsubscribe(interface{})
	GetAllSubscribedUsers() []int
	GetAllSubscribedGroups() []int64
}
