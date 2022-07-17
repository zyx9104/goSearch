package model

type WordMap struct {
	Map map[uint64]float32
	Len int
}

type ResponseDoc struct {
	Id    uint32  `json:"id,omitempty"`
	Text  string  `json:"text,omitempty"`
	Url   string  `json:"url,omitempty"`
	Score float32 `json:"score,omitempty"` //得分

}

// type ResponseUrl struct {
// 	ThumbnailUrl string  `json:"thumbnailUrl,omitempty"`
// 	Url          string  `json:"url,omitempty"`
// 	ID           uint64  `json:"id,omitempty"`
// 	Text         string  `json:"text,omitempty"`
// 	Score        float32 `json:"score,omitempty"`
// }

type RemoveIndexModel struct {
	ID uint64 `json:"id,omitempty"`
}
