package main

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

func (events *Events[Notification, Payment, T]) Match() (*Events[Notification, Payment, T], []*RelatedEvents[Notification, Payment]) {
	var relatedEvents []*RelatedEvents[Notification, Payment]

	var processedNotificationIndexes []int
	var processedPaymentIndexes []int
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
		}
	}

	return &Events[Notification, Payment, T]{
			Notifications:                         filteredNotifications,
			Payments:                              filteredPayments,
			ExpireCondition:                       events.ExpireCondition,
			NotificationWithPaymentMatchCondition: events.NotificationWithPaymentMatchCondition,
			PaymentsMatchCondition:                events.PaymentsMatchCondition,
		},
		relatedEvents

}

func Contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
