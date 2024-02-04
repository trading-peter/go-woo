package ws

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var (
	privateWsUrl = "wss://wss.woo.org/v2/ws/private/stream/%s"
)

type PrivateStream struct {
	appId                  string
	apiKey                 string
	apiSecret              string
	mu                     *sync.Mutex
	url                    string
	dialer                 *websocket.Dialer
	wsReconnectionCount    int
	wsReconnectionInterval time.Duration
	wsTimeout              time.Duration
	isDebugMode            bool
}

func NewPrivateStream(applicationId, apiKey, apiSecret string) *PrivateStream {
	return &PrivateStream{
		apiKey:                 apiKey,
		apiSecret:              apiSecret,
		appId:                  applicationId,
		url:                    fmt.Sprintf(privateWsUrl, applicationId),
		dialer:                 websocket.DefaultDialer,
		wsReconnectionCount:    reconnectCount,
		wsReconnectionInterval: reconnectInterval,
		wsTimeout:              streamTimeout,

		isDebugMode: true,
	}
}

func (s *PrivateStream) generateSignatureV3(timestamp string) string {
	signString := fmt.Sprintf("|%s", timestamp)

	h := hmac.New(sha256.New, []byte(s.apiSecret))
	h.Write([]byte(signString))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *PrivateStream) SubBalance(ctx context.Context) (<-chan BalanceEvent, error) {
	requests := []WSPrivateRequest{}

	requests = append(requests, WSPrivateRequest{
		ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		Event: "subscribe",
		Topic: "balance",
	})

	eventsC, err := servePrivate[BalanceEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan BalanceEvent, 1)
	go func() {
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-eventsC:
				if !ok {
					return
				}
				events <- event
			}
		}
	}()

	return events, nil
}

func (s *PrivateStream) printf(format string, v ...interface{}) {
	if !s.isDebugMode {
		return
	}
	log.Printf(format+"\n", v)
}

func (s *PrivateStream) connect(requests ...WSPrivateRequest) (*websocket.Conn, error) {
	conn, _, err := s.dialer.Dial(s.url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	s.printf("connected to %v", s.url)

	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	if err := conn.WriteJSON(&WSPrivateRequest{
		ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		Event: "auth",
		Params: RequestParams{
			Apikey:    s.apiKey,
			Sign:      s.generateSignatureV3(timestamp),
			Timestamp: timestamp,
		},
	}); err != nil {
		return nil, err
	}

	// wait to be auth...
	time.Sleep(time.Second)

	err = s.subscribe(conn, requests)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	conn.SetPongHandler(func(msg string) error {
		s.printf("%s", "PONG")
		err := conn.SetReadDeadline(time.Now().Add(s.wsTimeout))
		if err != nil {
			return err
		}
		return nil
	})

	return conn, nil
}

func servePrivate[T WSEventI](ctx context.Context, s *PrivateStream, requests ...WSPrivateRequest) (chan T, error) {
	topics := map[string]bool{}

	for _, w := range requests {
		topics[w.Topic] = true
	}

	conn, err := s.connect(requests...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	doneC := make(chan struct{})
	eventsC := make(chan T, 1)

	go func() {
		go func() {
			defer close(doneC)

			for {
				message := new(T)
				err = conn.ReadJSON(&message)

				if err != nil {
					s.printf("read msg: %v", err)

					/*
						if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
							return
						}
					*/

					conn, err = s.reconnect(ctx, requests)
					if err != nil {
						s.printf("reconnect: %+v", err)
						return
					}
					continue
				}

				msg := *message

				if topics[msg.GetTopic()] {
					eventsC <- msg
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					s.printf("write close msg: %v", err)
					return
				}
				select {
				case <-doneC:
					return
				case <-time.After(time.Second):
					return
				}
			case <-doneC:
				return
			case <-time.After(time.Second * 9):
				s.printf("%s", "PING")
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					s.printf("write ping: %v", err)
				}
			}
		}
	}()

	return eventsC, nil
}

func (s *PrivateStream) reconnect(ctx context.Context, requests []WSPrivateRequest) (*websocket.Conn, error) {
	for i := 1; i < s.wsReconnectionCount; i++ {
		conn, err := s.connect(requests...)
		if err == nil {
			return conn, nil
		}

		select {
		case <-time.After(s.wsReconnectionInterval):
			conn, err := s.connect(requests...)
			if err != nil {
				continue
			}

			return conn, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, errors.New("reconnection failed")
}

func (s *PrivateStream) subscribe(conn *websocket.Conn, requests []WSPrivateRequest) error {
	for _, req := range requests {
		err := conn.WriteJSON(req)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
