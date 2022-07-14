package tools

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"
	"fmt"
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

func Str2Uint64(str string) uint64 {
	hash := murmur3.Sum64([]byte(str))
	return hash
}

func U32ToBytes(key uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, key)
	return buf
}

func U64ToBytes(key uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, key)
	return buf
}

//ReadCsv return [[url, text], ...]
func ReadCsv(filepath string) [][]string {
	opencast, err := os.Open(filepath)
	if err != nil {
		logger.Panicln(filepath, err)
	}
	defer opencast.Close()

	logger.Debug("load:", filepath)
	//创建csv读取接口实例
	ReadCsv := csv.NewReader(opencast)
	// ReadCsv.Comma = ' '
	// ReadCsv.FieldsPerRecord = -1
	ReadCsv.Read()
	//读取所有内容
	ReadAll, err := ReadCsv.ReadAll() //返回切片类型：[[s s ds] [a a a]]
	if err != nil {
		logger.Panicln(filepath, err)
	}
	return ReadAll
}

func parseData(line []string, docs []model.Doc) []model.Doc {

	h1 := Str2Uint64(line[0])
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

//Encode 压缩[]byte
func Encode(input []byte) ([]byte, error) {
	// 创建一个新的 byte 输出流
	var buf bytes.Buffer
	// 创建一个新的 gzip 输出流
	gzipWriter := gzip.NewWriter(&buf)
	// 将 input byte 数组写入到此输出流中
	_, err := gzipWriter.Write(input)
	if err != nil {
		_ = gzipWriter.Close()
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	// 返回压缩后的 bytes 数组
	return buf.Bytes(), nil
}

//Decode 解压[]byte
func Decode(input []byte) ([]byte, error) {
	// 创建一个新的 gzip.Reader
	bytesReader := bytes.NewReader(input)
	gzipReader, err := gzip.NewReader(bytesReader)
	if err != nil {
		return nil, err
	}
	defer func() {
		// defer 中关闭 gzipReader
		_ = gzipReader.Close()
	}()
	buf := new(bytes.Buffer)
	// 从 Reader 中读取出数据
	if _, err := buf.ReadFrom(gzipReader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func HandleError(err error) {
	if err != nil {
		logger.Panic(err)
	}
}
