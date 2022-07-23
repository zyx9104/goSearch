package model

type ResponseDoc struct {
	Id    uint32  `json:"id,omitempty"`
	Text  string  `json:"text,omitempty"`
	Url   string  `json:"url,omitempty"`
	Score float64 `json:"score,omitempty"` //得分

}
