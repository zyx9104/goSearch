package main

import (
	"bufio"
	"os"
	"time"

	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/tools"
)

func init() {
	tools.Init()
	engine.Init()
}

func listen() {
	e := engine.DefaultEngine()
	e.Wait()

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		q := sc.Text()
		logger.Info("Start Search")
		filterWord := []string{"小孩", "儿童"}
		t := time.Now()
		docs := e.Search(q)
		st := time.Since(t)
		t = time.Now()
		docs = e.FliterResult(docs, filterWord)
		ft := time.Since(t)
		for i, doc := range docs {
			if i >= 10 {
				break
			}
			logger.Infoln(doc.Id, doc.Text, (len(doc.Text)+len(doc.Url)+4)/4)
			// fmt.Println(doc.Id, doc.Text)

		}
		logger.Infoln("Find docs:", len(docs), "Search time:", st, "Filter time:", ft)
		logger.Info("Search Done")
	}
	// gen.GenWordIds()
}

func main() {
	logger.Infoln("========================== Process Start ==========================")

	// s, _ := strconv.Atoi(os.Args[1])
	// e, _ := strconv.Atoi(os.Args[2])
	// // s, e := 1, viper.GetInt("db.last_index")
	// id, _ := strconv.Atoi(os.Args[3])
	listen()
	logger.Infoln("========================== Process Done ==========================")
}
