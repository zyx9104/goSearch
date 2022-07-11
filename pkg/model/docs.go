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

func (*Doc) TableName() string {
	return "docs"
}
