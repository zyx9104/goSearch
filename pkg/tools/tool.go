package tools

import (
	"bytes"
	"compress/flate"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/spaolacci/murmur3"
	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Db       *gorm.DB
	username string
	password string
	host     string
	port     int
	database string
	args     string
)

func init() {
	username = viper.GetString("db.username")
	password = viper.GetString("db.password")
	host = viper.GetString("db.host")
	port = viper.GetInt("db.port")
	database = viper.GetString("db.database")
	args = viper.GetString("db.args")
}

func StrTo64(str string) uint64 {
	hash := murmur3.Sum64([]byte(str))
	return hash
}

func ReadCsv(filepath string) [][]string {
	opencast, err := os.Open(filepath)
	if err != nil {
		log.Println("csv文件打开失败！")
	}
	defer opencast.Close()

	logger.Debug("load:", filepath)
	//创建csv读取接口实例
	ReadCsv := csv.NewReader(opencast)
	ReadCsv.Read()
	//读取所有内容
	ReadAll, err := ReadCsv.ReadAll() //返回切片类型：[[s s ds] [a a a]]
	if err != nil {
		logger.Log.Fatal(err)
	}
	return ReadAll
}

func parseData(line []string, docs []model.Doc) []model.Doc {

	h1 := StrTo64(line[0])
	docs = append(docs, model.Doc{Hash: h1, Url: line[0], Text: line[1]})
	if len(docs) >= 5000 {
		rs := Db.Create(&docs)
		if rs.Error != nil {
			logger.Log.Debug(rs.Error)
		}
		logger.Log.Debugf("load %d lines data:", rs.RowsAffected)
		docs = docs[:0]
	}
	return docs
}

func ReadToDb() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", username, password, host, port, database, args)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Doc{})
	var filepath string

	for i := 0; i <= 255; i++ {
		filepath = fmt.Sprintf("../wukong_release/wukong_100m_%d.csv", i)

		ReadAll := ReadCsv(filepath)
		docs := make([]model.Doc, 0)

		for _, line := range ReadAll {
			docs = parseData(line, docs)
		}
		db.Create(&docs)
		logger.Log.Debugf("load %d lines data:", len(docs))
	}

}

func ExecTime(fn func()) float64 {
	start := time.Now()
	fn()
	tc := float64(time.Since(start).Nanoseconds())
	return tc / 1e6
}

// Write 写入二进制数据到磁盘文件
func Write(data interface{}, filename string) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		logger.Panic(err)
	}

	logger.Debug("Write:", filename)
	compressData := Compression(buffer.Bytes())
	err = ioutil.WriteFile(filename, compressData, 0600)
	if err != nil {
		logger.Panic(err)
	}
}

func Encoder(data interface{}) []byte {
	if data == nil {
		return nil
	}
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func Decoder(data []byte, v interface{}) {
	if data == nil {
		return
	}
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(v)
	if err != nil {
		panic(err)
	}
}

// Compression 压缩数据
func Compression(data []byte) []byte {

	buf := new(bytes.Buffer)
	write, err := flate.NewWriter(buf, flate.DefaultCompression)
	if err != nil {
		logger.Panic(err)
	}
	defer write.Close()

	write.Write(data)
	write.Flush()
	logger.Debug("原大小：", len(data), "压缩后大小：", buf.Len(), "压缩率：", fmt.Sprintf("%.2f", float32(buf.Len())*100/float32(len(data))), "%")
	return buf.Bytes()
}

//Decompression 解压缩数据
func Decompression(data []byte) []byte {
	return DecompressionBuffer(data).Bytes()
}

func DecompressionBuffer(data []byte) *bytes.Buffer {
	buf := new(bytes.Buffer)
	read := flate.NewReader(bytes.NewReader(data))
	defer read.Close()

	buf.ReadFrom(read)
	return buf
}

// Read 从磁盘文件加载二进制数据
func Read(data interface{}, filename string) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			//忽略
			return
		}
		logger.Debug(err)
	}
	//解压
	decoData := Decompression(raw)

	buffer := bytes.NewBuffer(decoData)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(data)
	if err != nil {
		logger.Debug("Decode Error: ", err, "buffer.Bytes() is :", buffer.Bytes())
	}
}
