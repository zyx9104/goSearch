package model

import (
	"gorm.io/gorm"
)

type Doc struct {
	gorm.Model
	//FavoriteId uint
	Hash uint64
	Url  string
	Text string
}

type ResponseDoc struct {
	Id    uint32  `json:"id,omitempty"`
	Text  string  `json:"text,omitempty"`
	Url   string  `json:"url,omitempty"`
	Score float64 `json:"score,omitempty"` //得分

}
