package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationExpired(t *testing.T) {
	notification1 := &Pair[*int, *int]{First: IntRef(1), Second: IntRef(1)}
	notification2 := &Pair[*int, *int]{First: IntRef(2), Second: IntRef(2)}

	events := &Events[int, int, int]{
		Notifications:   []*Pair[*int, *int]{notification1, notification2},
		ExpireCondition: func(t *int) bool { return *t < 2 },
	}

	newEvents, _ := events.Match()

	assert.Equal(t, 1, len(newEvents.Notifications))
	assert.Equal(t, 2, *newEvents.Notifications[0].First)
}

func TestMatchOfExpiredOccured(t *testing.T) {
	notification1 := &Pair[*int, *int]{First: IntRef(1), Second: IntRef(1)}
	notification2 := &Pair[*int, *int]{First: IntRef(2), Second: IntRef(2)}

	payment1 := &Pair[*int, *int]{First: IntRef(1), Second: IntRef(1)}
	payment2 := &Pair[*int, *int]{First: IntRef(2), Second: IntRef(2)}

	events := &Events[int, int, int]{
		Notifications:   []*Pair[*int, *int]{notification1, notification2},
		Payments:        []*Pair[*int, *int]{payment1, payment2},
		ExpireCondition: func(t *int) bool { return *t < 2 },
		NotificationWithPaymentMatchCondition: func(n, p *int) bool {
			return *n == *p
		},
	}

	newEvents, relatedEvents := events.Match()

	//Check that elements with 1 removed from list and matched
	//and elements with 2 are still there
	assert.Equal(t, 1, len(newEvents.Notifications))
	assert.Equal(t, 1, len(newEvents.Payments))

	assert.Equal(t, 2, *newEvents.Notifications[0].First)
	assert.Equal(t, 2, *newEvents.Payments[0].First)

	assert.Equal(t, 1, len(relatedEvents))

	assert.Equal(t, 1, *relatedEvents[0].Notification)
	assert.Equal(t, 1, len(relatedEvents[0].Payments))
	assert.Equal(t, 1, *relatedEvents[0].Payments[0])
}

func TestExpiredAreRemoved(t *testing.T) {
	notification1 := &Pair[*int, *int]{First: IntRef(1), Second: IntRef(1)}
	notification2 := &Pair[*int, *int]{First: IntRef(2), Second: IntRef(2)}

	events := &Events[int, int, int]{
		Notifications:   []*Pair[*int, *int]{notification1, notification2},
		ExpireCondition: func(t *int) bool { return *t < 3 },
	}

	newEvents, relatedEvents := events.Match()

	assert.Equal(t, 0, len(newEvents.Notifications))
	assert.Equal(t, 0, len(relatedEvents))
}
