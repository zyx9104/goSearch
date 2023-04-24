package tools

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spaolacci/murmur3"
	"github.com/spf13/viper"
	"github.com/wangbin/jiebago"

	log1 "github.com/z-y-x233/goSearch/pkg/log"
)

var (
	Seg     jiebago.Segmenter
	WordIds map[uint64]int
)

func Init() {
	viper.SetConfigFile("./config.yml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("load config failed:", err)
	}
	err = log1.Init()
	if err != nil {
		log.Fatal("init logger failed:", err)
	}

	log1.Infoln("Load dict")
	dictPath := viper.GetString("db.dict")
	err = Seg.LoadDictionary(dictPath)
	HandleError(fmt.Sprintf("load %s failed:", dictPath), err)

	log1.Infoln("Load Word Ids")
	WordIds = make(map[uint64]int, 16000000)
	wordIdsPath := viper.GetString("db.wordIds")
	LoadWordIds(wordIdsPath)
}

func LoadWordIds(path string) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0664)
	HandleError("load word ids failed:", err)
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = ' '
	lines, err := reader.ReadAll()
	HandleError("read failed:", err)
	for _, line := range lines {
		word, _ := strconv.ParseUint(line[0], 10, 0)
		ids, _ := strconv.Atoi(line[1])
		WordIds[word] = ids
	}
}

func getWords(chh <-chan string) (words []string) {

	for word := range chh {
		words = append(words, word)
	}
	return
}

func WordCutForInv(q string) []string {
	//不区分大小写
	q = strings.ToLower(q)
	//移除所有的标点符号
	q = RemovePunctuation(q)
	//移除所有的空格
	q = RemoveSpace(q)
	ch := Seg.CutForSearch(q, true)
	words := getWords(ch)

	return words
}

func WordCut(q string) (res []string) {

	//不区分大小写
	q = strings.ToLower(q)
	//移除所有的标点符号
	q = RemovePunctuation(q)

	ch := Seg.CutForSearch(q, true)
	words := getWords(ch)
	wordSet := make(map[string]int, 10)
	length := 0
	for _, word := range words {
		wordSet[word]++
		if wordSet[word] == 1 {
			key := Str2Uint64(word)
			length += WordIds[key]
		}
	}
	// logger.Info(wordSet)
	for word := range wordSet {
		key := Str2Uint64(word)
		// logger.Infoln(word, len([]rune(word)), WordIds[key])
		if len([]rune(word)) == 1 && WordIds[key] >= 10000 {
			continue
		}
		if len(words) >= 5 && WordIds[key]*2 >= length {
			continue
		}
		res = append(res, word)
	}
	if len(res) == 0 && len(words) > 0 {
		res = append(res, words[0])
	}
	return res
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

// ReadCsv return [[url, text], ...]
func ReadCsv(filepath string) [][]string {
	opencast, err := os.Open(filepath)
	if err != nil {
		log1.Panicln(filepath, err)
	}
	defer opencast.Close()

	log1.Debug("load:", filepath)
	ReadCsv := csv.NewReader(opencast)
	// ReadCsv.Comma = ' '
	// ReadCsv.FieldsPerRecord = -1
	//读取第一行
	ReadCsv.Read()
	ReadAll, err := ReadCsv.ReadAll() //返回切片类型：[[s s ds] [a a a]]
	if err != nil {
		log1.Panicln(filepath, err)
	}
	return ReadAll
}

func HandleError(msg string, err error) {
	if err != nil {
		log1.Panicln(msg, err)
	}
}

func Set(ids []uint32) []uint32 {
	ht := make(map[uint32]bool, len(ids))
	for _, id := range ids {
		ht[id] = true
	}
	ids = ids[:0]
	for k := range ht {
		ids = append(ids, k)
	}
	return ids
}

// RemovePunctuation 移除所有的标点符号
func RemovePunctuation(str string) string {
	reg := regexp.MustCompile(`\p{P}+`)
	return reg.ReplaceAllString(str, "")
}

// RemoveSpace 移除所有的空格
func RemoveSpace(str string) string {
	reg := regexp.MustCompile(`\s+`)
	return reg.ReplaceAllString(str, "")
}

func IDF(word uint64, N int) float64 {
	return math.Log(float64(N+1)) / math.Log(float64(float64(WordIds[word])+0.5))
}

func R(tf int) float64 {
	var (
		k1 float64 = 1.2
		b  float64 = 0.75
	)
	up := k1 * float64(tf)
	down := k1*(1-b) + float64(tf)
	return up / down
}

func WriteBytes(data []byte, filename string) error {
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	log1.Debugln("write", len(data), "bytes to", filename)
	return nil
}

func ReadBytes(filename string) (data []byte, err error) {

	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	log1.Debugln("read", len(data), "bytes from", filename)
	return data, nil
}
