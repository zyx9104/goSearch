package model

type DocIndex struct {
	ID   uint32 `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
}

// StorageIndexDoc 文档对象
type StorageIndexDoc struct {
	Text string `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
}

// StorageId leveldb中的Ids存储对象
type StorageId struct {
	ID    uint32
	Score float32
}

type InvIndex struct {
	ID []uint32
}

type WordMap struct {
	Map map[uint64]float32
	Len int
}

type ResponseDoc struct {
	DocIndex
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
