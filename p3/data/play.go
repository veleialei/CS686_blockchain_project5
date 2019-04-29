package data

type PlayData struct {
	Id     int32  `json:"id"`
	Addr   string `json:"addr"`
	Height int32  `json:"height"`
	Hash   string `json:"hash"`
	React  string `json:"react"`
}
