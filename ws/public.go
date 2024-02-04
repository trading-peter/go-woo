package ws

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	publicWsUrl = "wss://wss.woo.org/ws/stream/%s"

	writeWait         = time.Second * 10
	reconnectCount    = int(10)
	reconnectInterval = time.Second
	streamTimeout     = time.Second * 60
)

type PublicStream struct {
	appId                  string
	mu                     *sync.Mutex
	url                    string
	dialer                 *websocket.Dialer
	wsReconnectionCount    int
	wsReconnectionInterval time.Duration
	wsTimeout              time.Duration
	isDebugMode            bool
}

func NewPublicStream(applicationId string) *PublicStream {
	return &PublicStream{
		appId:                  applicationId,
		url:                    fmt.Sprintf(publicWsUrl, applicationId),
		dialer:                 websocket.DefaultDialer,
		wsReconnectionCount:    reconnectCount,
		wsReconnectionInterval: reconnectInterval,
		wsTimeout:              streamTimeout,

		isDebugMode: true,
	}
}

func (s *PublicStream) SubOrderBook(ctx context.Context, is100 bool, symbols ...string) (<-chan OrderbookEvent, error) {
	requests := []WSRequest{}

	topic := "orderbook"
	if is100 {
		topic += "100"
	}

	for _, sym := range symbols {
		requests = append(requests, WSRequest{
			ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
			Event: "subscribe",
			Topic: fmt.Sprintf("%s@%s", sym, topic),
		})
	}

	eventsC, err := serve[OrderbookEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan OrderbookEvent, 1)
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

func (s *PublicStream) SubBestBookOffer(ctx context.Context, symbols ...string) (<-chan BestBookOfferEvent, error) {
	requests := []WSRequest{}

	for _, sym := range symbols {
		requests = append(requests, WSRequest{
			ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
			Event: "subscribe",
			Topic: fmt.Sprintf("%s@bbo", sym),
		})
	}

	eventsC, err := serve[BestBookOfferEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan BestBookOfferEvent, 1)
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

func (s *PublicStream) SubBestBooksOffers(ctx context.Context) (<-chan BestBooksOffersEvent, error) {
	requests := []WSRequest{}

	requests = append(requests, WSRequest{
		ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		Event: "subscribe",
		Topic: "bbos",
	})

	eventsC, err := serve[BestBooksOffersEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan BestBooksOffersEvent, 1)
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

func (s *PublicStream) SubTrade(ctx context.Context, symbols ...string) (<-chan TradeEvent, error) {
	requests := []WSRequest{}

	for _, sym := range symbols {
		requests = append(requests, WSRequest{
			ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
			Event: "subscribe",
			Topic: fmt.Sprintf("%s@trade", sym),
		})
	}

	eventsC, err := serve[TradeEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan TradeEvent, 1)
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

func (s *PublicStream) SubOpenInterest(ctx context.Context) (<-chan OpenInterestEvent, error) {
	requests := []WSRequest{}

	requests = append(requests, WSRequest{
		ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		Event: "subscribe",
		Topic: "openinterests",
	})

	eventsC, err := serve[OpenInterestEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan OpenInterestEvent, 1)
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

func (s *PublicStream) SubEstFundingRate(ctx context.Context, symbols ...string) (<-chan EstFundingRateEvent, error) {
	requests := []WSRequest{}

	for _, sym := range symbols {
		requests = append(requests, WSRequest{
			ID:    fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
			Event: "subscribe",
			Topic: fmt.Sprintf("%s@estfundingrate", sym),
		})
	}

	eventsC, err := serve[EstFundingRateEvent](ctx, s, requests...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	events := make(chan EstFundingRateEvent, 1)
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

func (s *PublicStream) printf(format string, v ...interface{}) {
	if !s.isDebugMode {
		return
	}
	log.Printf(format+"\n", v)
}

func (s *PublicStream) connect(requests ...WSRequest) (*websocket.Conn, error) {
	conn, _, err := s.dialer.Dial(s.url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	s.printf("connected to %v", s.url)

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

func serve[T WSEventI](ctx context.Context, s *PublicStream, requests ...WSRequest) (chan T, error) {
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

func (s *PublicStream) reconnect(ctx context.Context, requests []WSRequest) (*websocket.Conn, error) {
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

func (s *PublicStream) subscribe(conn *websocket.Conn, requests []WSRequest) error {
	for _, req := range requests {
		err := conn.WriteJSON(req)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
