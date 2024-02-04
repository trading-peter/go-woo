package ws

type WSEventI interface {
	GetTopic() string
}

type WSRequest struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Topic string `json:"topic"`
}

type BestBookOfferEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  struct {
		Symbol  string  `json:"symbol"`
		Ask     float64 `json:"ask"`
		AskSize float64 `json:"askSize"`
		Bid     float64 `json:"bid"`
		BidSize float64 `json:"bidSize"`
	} `json:"data"`
}

func (e BestBookOfferEvent) GetTopic() string {
	return e.Topic
}

type TradeEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"price"`
		Size   float64 `json:"size"`
		Side   string  `json:"side"`
		Source int     `json:"source"`
	} `json:"data"`
}

func (e TradeEvent) GetTopic() string {
	return e.Topic
}

type BestBooksOffersEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  []struct {
		Symbol  string  `json:"symbol"`
		Ask     float64 `json:"ask"`
		AskSize float64 `json:"askSize"`
		Bid     float64 `json:"bid"`
		BidSize float64 `json:"bidSize"`
	} `json:"data"`
}

func (e BestBooksOffersEvent) GetTopic() string {
	return e.Topic
}

type OpenInterestEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  []struct {
		Symbol       string  `json:"symbol"`
		Openinterest float64 `json:"openinterest"`
	} `json:"data"`
}

func (e OpenInterestEvent) GetTopic() string {
	return e.Topic
}

type EstFundingRateEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  struct {
		Symbol      string  `json:"symbol"`
		FundingRate float64 `json:"fundingRate"`
		FundingTs   int64   `json:"fundingTs"`
	} `json:"data"`
}

func (e EstFundingRateEvent) GetTopic() string {
	return e.Topic
}

type OrderbookEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  struct {
		Symbol string      `json:"symbol"`
		Asks   [][]float64 `json:"asks"`
		Bids   [][]float64 `json:"bids"`
	} `json:"data"`
}

func (e OrderbookEvent) GetTopic() string {
	return e.Topic
}
