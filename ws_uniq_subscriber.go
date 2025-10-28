package pacifica

import (
	"sync"
)

type callback func(any)

type uniqSubscriber struct {
	mu                  sync.Mutex
	id                  string
	count               int64
	subscribers         map[string]callback
	subscriberFunc      func(subscriptable)
	unsubscriberFunc    func(subscriptable)
	subscriptionPayload subscriptable
}

func newUniqSubscriber(
	id string,
	payload subscriptable,
	subscriberFunc, unsubscriberFunc func(subscriptable),
) *uniqSubscriber {
	return &uniqSubscriber{
		id:                  id,
		subscriptionPayload: payload,
		count:               0,
		subscribers:         make(map[string]callback),
		subscriberFunc:      subscriberFunc,
		unsubscriberFunc:    unsubscriberFunc,
	}
}

func (u *uniqSubscriber) subscribe(id string, cb callback) {
	u.mu.Lock()
	if _, exists := u.subscribers[id]; exists {
		u.mu.Unlock()
		return
	}
	u.subscribers[id] = cb
	u.count++
	c := u.count
	u.mu.Unlock()

	if c == 1 {
		u.subscriberFunc(u.subscriptionPayload)
	}
}

func (u *uniqSubscriber) unsubscribe(id string) {
	u.mu.Lock()
	if _, exists := u.subscribers[id]; !exists {
		u.mu.Unlock()
		return
	}
	delete(u.subscribers, id)
	c := u.count - 1
	u.count = c
	u.mu.Unlock()

	if c == 0 {
		u.unsubscriberFunc(u.subscriptionPayload)
	}
}

func (u *uniqSubscriber) dispatch(data any) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for _, cb := range u.subscribers {
		cb(data)
	}
}

func (u *uniqSubscriber) clear() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.subscribers = make(map[string]callback)
	u.count = 0
	u.unsubscriberFunc(u.subscriptionPayload)
}
