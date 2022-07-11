package tools

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/spaolacci/murmur3"
	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func StrTo64(str string) uint64 {
	hash := murmur3.Sum64([]byte(str))
	return hash
}

func ReadToDb() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		viper.GetString("db.username"), viper.GetString("db.password"),
		viper.GetString("db.host"), viper.GetInt("db.port"),
		viper.GetString("db.database"), viper.GetString("db.args"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Doc{})
	var filepath string

	for i := 0; i <= 255; i++ {
		filepath = fmt.Sprintf("./data/wukong_release/wukong_100m_%d.csv", i)
		opencast, err := os.Open(filepath)
		if err != nil {
			log.Println("csv文件打开失败！")
		}
		defer opencast.Close()

		logger.Log.Debug("load:", filepath)
		//创建csv读取接口实例
		ReadCsv := csv.NewReader(opencast)
		ReadCsv.Read()
		//读取所有内容
		ReadAll, err := ReadCsv.ReadAll() //返回切片类型：[[s s ds] [a a a]]
		if err != nil {
			logger.Log.Fatal(err)
		}
		docs := make([]model.Doc, 0)

		for _, line := range ReadAll {
			h1, h2 := StrTo64(line[0])
			docs = append(docs, model.Doc{H1: h1, H2: h2, Url: line[0], Text: line[1]})
			if len(docs) >= 5000 {
				rs := db.Create(&docs)
				if rs.Error != nil {
					logger.Log.Debug(rs.Error)
				}
				logger.Log.Debugf("load %d lines data:", rs.RowsAffected)
				docs = docs[:0]
			}
			//logger.Log.Debug(line)
		}
		logger.Log.Debugf("load %d lines data:", len(docs))
		db.Create(&docs)
	}

	// r := db.Find(&docs)
	// logger.Infoln(r)
	// for i, l := range docs {
	// 	logger.Log.Debug(l)
	// 	if i == 10 {
	// 		break
	// 	}
	// }
}
