package pacifica

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sonirico/vago/maps"
)

const (
	// pingInterval is the interval for sending ping messages to keep WebSocket alive
	pingInterval = 50 * time.Second
)

type logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

type WebsocketClient struct {
	url                   string
	done                  chan struct{}
	conn                  *websocket.Conn
	closeOnce             sync.Once
	reconnectWait         time.Duration
	mu                    sync.RWMutex
	writeMu               sync.Mutex
	subscribers           map[string]*uniqSubscriber
	msgDispatcherRegistry map[string]msgDispatcher
	logger                logger
	nextSubID             atomic.Int64

	debug bool
}

type Subscription struct {
	ID      string
	Payload any
	Close   func()
}

func NewWebsocketClient(url string, opts ...WsOpt) *WebsocketClient {
	if url == "" {
		url = MainnetWSURL
	}
	client := &WebsocketClient{
		url:           url,
		reconnectWait: time.Second,
		done:          make(chan struct{}),
		subscribers:   make(map[string]*uniqSubscriber),
		msgDispatcherRegistry: map[string]msgDispatcher{
			ChannelOrderBook: newMsgDispatcher[OrderBook](ChannelOrderBook),
		},
	}

	for _, opt := range opts {
		opt.Apply(client)
	}

	return client
}

func (w *WebsocketClient) Connect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		return nil
	}

	dialer := websocket.Dialer{}

	conn, _, err := dialer.DialContext(ctx, w.url, nil)
	if err != nil {
		return err
	}

	w.conn = conn

	go w.pingPump(ctx)
	go w.readPump(ctx)

	return w.resubscribeAll()
}

func (w *WebsocketClient) Close() error {
	var err error
	w.closeOnce.Do(func() {
		err = w.close()
	})
	return err
}

func (w *WebsocketClient) close() error {
	close(w.done)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		return w.conn.Close()
	}

	for _, subscriber := range w.subscribers {
		subscriber.clear()
	}

	return nil
}

func (w *WebsocketClient) resubscribeAll() error {
	for _, subscriber := range w.subscribers {
		if err := w.sendSubscribe(subscriber.subscriptionPayload); err != nil {
			return fmt.Errorf("resubscribe: %w", err)
		}
	}
	return nil
}

func (w *WebsocketClient) sendSubscribe(payload any) error {
	return w.writeJSON(wsCommand{
		Method: "subscribe",
		Params: payload,
	})
}

func (w *WebsocketClient) sendUnsubscribe(payload any) error {
	return w.writeJSON(wsCommand{
		Method: "unsubscribe",
		Params: payload,
	})
}

func (w *WebsocketClient) sendPing() error {
	return w.writeJSON(wsCommand{Method: "ping"})
}

func (w *WebsocketClient) writeJSON(v any) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	if w.conn == nil {
		return errors.New("connection closed")
	}

	if w.debug {
		bts, _ := json.Marshal(v)
		w.logDebugf("[>] %s", string(bts))
	}

	return w.conn.WriteJSON(v)
}

func (w *WebsocketClient) pingPump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
			if err := w.sendPing(); err != nil {
				w.logErrf("failed to send ping: %v", err)
				w.reconnect(ctx)
				return
			}
		}
	}
}

func (w *WebsocketClient) readPump(ctx context.Context) {
	defer func() {
		w.mu.Lock()
		if w.conn != nil {
			_ = w.conn.Close()
			w.conn = nil
		}
		w.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		default:
			_, msg, err := w.conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					w.logErrf("websocket read error: %v", err)
				}
				return
			}

			if w.debug {
				w.logDebugf("[<] %s", string(msg))
			}

			var wsMsg wsMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				w.logErrf("websocket message parse error: %v", err)
				continue
			}

			if err := w.dispatch(wsMsg); err != nil {
				w.logErrf("failed to dispatch websocket message: %v", err)
			}
		}
	}
}

func (w *WebsocketClient) dispatch(msg wsMessage) error {
	dispatcher, ok := w.msgDispatcherRegistry[msg.Channel]
	if !ok {
		return fmt.Errorf("no dispatcher for channel: %s", msg.Channel)
	}

	w.mu.Lock()
	subscribers := maps.Values(w.subscribers)
	w.mu.Unlock()

	return dispatcher.Dispatch(subscribers, msg)
}

func (w *WebsocketClient) reconnect(ctx context.Context) {
	for {
		select {
		case <-w.done:
			return
		case <-ctx.Done():
			return
		default:
			if err := w.Connect(ctx); err == nil {
				return
			}
			time.Sleep(w.reconnectWait)
			w.reconnectWait *= 2
			if w.reconnectWait > time.Minute {
				w.reconnectWait = time.Minute
			}
		}
	}
}

func (w *WebsocketClient) logErrf(fmt string, args ...any) {
	if w.logger == nil {
		return
	}

	w.logger.Errorf(fmt, args...)
}

func (w *WebsocketClient) logDebugf(fmt string, args ...any) {
	if w.logger == nil {
		return
	}

	w.logger.Infof(fmt, args...)
}

func (w *WebsocketClient) subscribe(payload subscriptable, callback func(msg any)) (*Subscription, error) {
	if callback == nil {
		return nil, fmt.Errorf("callback cannot be nil")
	}

	w.mu.Lock()

	pKey := payload.Key()
	subscriber, exists := w.subscribers[pKey]
	if !exists {
		subscriber = newUniqSubscriber(
			pKey,
			payload,
			func(p subscriptable) {
				if err := w.sendSubscribe(p); err != nil {
					w.logErrf("failed to subscribe: %v", err)
				}
			},
			func(p subscriptable) {
				w.mu.Lock()
				defer w.mu.Unlock()
				delete(w.subscribers, pKey)
				if err := w.sendUnsubscribe(p); err != nil {
					w.logErrf("failed to unsubscribe: %v", err)
				}
			},
		)
		w.subscribers[pKey] = subscriber
	}

	w.mu.Unlock()

	nextID := w.nextSubID.Add(1)
	subID := key(pKey, strconv.Itoa(int(nextID)))
	subscriber.subscribe(subID, callback)
	return &Subscription{
		ID: subID,
		Close: func() {
			subscriber.unsubscribe(subID)
		},
	}, nil
}
