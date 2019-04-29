package data

type CreateData struct {
	Id           string `json:"id"`
	ParentHeight int32  `json:"parentHeight"`
	ParentHash   string `json:"parentHash"`
	Content      string `json:"hash"`
	React        string `json:"react"`
	Secret       string `json:"secret"`
}
