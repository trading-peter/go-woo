package ws

type WSPrivateRequest struct {
	ID     string        `json:"id"`
	Event  string        `json:"event"`
	Topic  string        `json:"topic"`
	Params RequestParams `json:"params"`
}

type RequestParams struct {
	Apikey    string `json:"apikey"`
	Sign      string `json:"sign"`
	Timestamp string `json:"timestamp"`
}

type BalanceEvent struct {
	Topic string `json:"topic"`
	Ts    int64  `json:"ts"`
	Data  struct {
		Balances map[string]struct {
			Holding          float64 `json:"holding"`
			Frozen           float64 `json:"frozen"`
			Interest         float64 `json:"interest"`
			PendingShortQty  float64 `json:"pendingShortQty"`
			PendingLongQty   float64 `json:"pendingLongQty"`
			Version          float64 `json:"version"`
			Staked           float64 `json:"staked"`
			Unbonding        float64 `json:"unbonding"`
			Vault            float64 `json:"vault"`
			LaunchpadVault   float64 `json:"launchpadVault"`
			Earn             float64 `json:"earn"`
			AverageOpenPrice float64 `json:"averageOpenPrice"`
			Pnl24H           float64 `json:"pnl24H"`
			Fee24H           float64 `json:"fee24H"`
			MarkPrice        float64 `json:"markPrice"`
		} `json:"balances"`
	} `json:"data"`
}

func (e BalanceEvent) GetTopic() string {
	return e.Topic
}
