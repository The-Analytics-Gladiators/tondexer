package core

import "time"

type Pair[K any, V any] struct {
	First  K
	Second V
}

type Events[Notification any, Payment any, T comparable] struct {
	Notifications                         []*Pair[*Notification, *T]
	Payments                              []*Pair[*Payment, *T]
	ExpireCondition                       func(t *T) bool
	NotificationWithPaymentMatchCondition func(n *Notification, p *Payment) bool
	PaymentsMatchCondition                func(p1 *Payment, p2 *Payment) bool
}

type RelatedEvents[Notification any, Payment any] struct {
	Notification *Notification
	Payments     []*Payment
}

type OrphanEvents[Notification any, Payment any] struct {
	Notifications []*Notification
	Payments      []*Payment
}

func (events *Events[Notification, Payment, T]) Match() (*Events[Notification, Payment, T], []*RelatedEvents[Notification, Payment], OrphanEvents[Notification, Payment]) {
	var relatedEvents []*RelatedEvents[Notification, Payment]

	var processedNotificationIndexes []int
	var processedPaymentIndexes []int

	var orphanEvents OrphanEvents[Notification, Payment]
	for notificationIndex, notification := range events.Notifications {
		if events.ExpireCondition(notification.Second) {
			processedNotificationIndexes = append(processedNotificationIndexes, notificationIndex)

			relatedEvent := &RelatedEvents[Notification, Payment]{Notification: notification.First}

			for paymentIndex, payment := range events.Payments {
				if events.NotificationWithPaymentMatchCondition(notification.First, payment.First) {
					processedPaymentIndexes = append(processedPaymentIndexes, paymentIndex)

					relatedEvent.Payments = append(relatedEvent.Payments, payment.First)
				}
			}

			if len(relatedEvent.Payments) != 0 {
				relatedEvents = append(relatedEvents, relatedEvent)
			} else {
				orphanEvents.Notifications = append(orphanEvents.Notifications, notification.First)
			}
		}
	}

	var filteredNotifications []*Pair[*Notification, *T]
	var filteredPayments []*Pair[*Payment, *T]

	for notificationIndex, notification := range events.Notifications {
		if !Contains(processedNotificationIndexes, notificationIndex) {
			filteredNotifications = append(filteredNotifications, notification)
		}
	}

	for paymentIndex, payment := range events.Payments {
		if !Contains(processedPaymentIndexes, paymentIndex) && !events.ExpireCondition(payment.Second) {
			filteredPayments = append(filteredPayments, payment)
		} else if events.ExpireCondition(payment.Second) && !Contains(processedPaymentIndexes, paymentIndex) {
			//log.Printf("Not matched payment %v \n", payment)
			orphanEvents.Payments = append(orphanEvents.Payments, payment.First)
		}
	}

	return &Events[Notification, Payment, T]{
			Notifications:                         filteredNotifications,
			Payments:                              filteredPayments,
			ExpireCondition:                       events.ExpireCondition,
			NotificationWithPaymentMatchCondition: events.NotificationWithPaymentMatchCondition,
			PaymentsMatchCondition:                events.PaymentsMatchCondition,
		},
		relatedEvents,
		orphanEvents
}

type WaitingList[T any] struct {
	Entities          []*Pair[T, time.Time]
	ExpirationSeconds time.Duration
}

func (waitingList *WaitingList[T]) Add(t T) {
	waitingList.Entities = append(waitingList.Entities, &Pair[T, time.Time]{t, time.Now()})
}

func (waitingList *WaitingList[T]) Evict() []T {
	var evicted []T
	var remained []*Pair[T, time.Time]

	for _, pair := range waitingList.Entities {
		if time.Now().After(pair.Second.Add(waitingList.ExpirationSeconds)) {
			evicted = append(evicted, pair.First)
		} else {
			remained = append(remained, pair)
		}
	}
	waitingList.Entities = remained
	return evicted
}

type EvictableSet[T comparable] struct {
	mp                map[T]time.Time
	ExpirationSeconds time.Duration
}

func NewEvictableSet[T comparable](expiration time.Duration) *EvictableSet[T] {
	return &EvictableSet[T]{
		mp:                map[T]time.Time{},
		ExpirationSeconds: expiration,
	}
}

func (set *EvictableSet[T]) Add(t T) time.Time {
	now := time.Now()
	set.mp[t] = now
	return now
}

func (set *EvictableSet[T]) Evict() []T {
	var evicted []T
	for k, v := range set.mp {
		if v.After(time.Now().Add(set.ExpirationSeconds)) {
			delete(set.mp, k)
			evicted = append(evicted, k)
		}
	}

	return evicted
}

func (set *EvictableSet[T]) Exists(t T) bool {
	_, exists := set.mp[t]
	return exists
}
