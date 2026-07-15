package session

type Token struct {
	Value     string `json:"value"`
	ExpiresAt int64  `json:"expires_at"`
}

func (t Token) Valid(now int64) bool {
	return t.Value != "" && now < t.ExpiresAt
}

type Data struct {
	ServerURL string `json:"server_url"`
	PlayerID  string `json:"player_id"`
	Session   Token  `json:"session"`
	Refresh   Token  `json:"refresh"`
	GameID    string `json:"game_id,omitempty"`
	GameToken string `json:"game_token,omitempty"`
}

type Store interface {
	Load() (Data, error)
	Save(data Data) error
	Clear() error
}
