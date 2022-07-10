package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func init() {
	viper.SetConfigFile("./config.json")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("load config failed:", err)
	}
	err = logger.Init()
	if err != nil {
		log.Fatalln("init logger failed:", err)

	}
}

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		viper.GetString("db.username"), viper.GetString("db.password"),
		viper.GetString("db.host"), viper.GetInt("db.port"),
		viper.GetString("db.database"), viper.GetString("db.args"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Doc{})

	//filepath := "./data/wukong_test.csv"
	//opencast, err := os.Open(filepath)
	//if err != nil {
	//	log.Println("csv文件打开失败！")
	//}
	//defer opencast.Close()
	//
	////创建csv读取接口实例
	//ReadCsv := csv.NewReader(opencast)
	//ReadCsv.Read()
	////读取所有内容
	//ReadAll, err := ReadCsv.ReadAll() //返回切片类型：[[s s ds] [a a a]]

	docs := make([]model.Doc, 0)

	//for _, line := range ReadAll {
	//	docs = append(docs, model.Doc{Summary: line[1], Url: line[3]})
	//	if len(docs) > 5000 {
	//		db.Create(&docs)
	//		docs = docs[:0]
	//	}
	//	//logger.Log.Debug(line)
	//}
	//db.Create(&docs)

	r := db.Find(&docs)
	logger.Infoln(r)
	for i, l := range docs {
		logger.Log.Debug(l)
		if i == 10 {
			break
		}
	}
}
