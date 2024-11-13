package core

import "time"

type Pair[K any, V any] struct {
	First  K
	Second V
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
		if time.Now().After(v.Add(set.ExpirationSeconds)) {
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
