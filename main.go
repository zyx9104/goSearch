package main

import (
	"bufio"
	"os"
	"time"

	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"github.com/z-y-x233/goSearch/pkg/tree"
)

func init() {
	tools.Init()
	engine.Init()
	tree.Init()
}

func Listen() {
	e := engine.DefaultEngine()
	e.Wait()

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		q := sc.Text()
		// q := "1"
		logger.Info("Start Search")
		t := time.Now()
		ss := tree.FindRelated(q, 10)
		st := time.Since(t)
		for _, s := range ss {
			logger.Info(s)
		}
		logger.Debugln("search time:", st)
		// filterWord := []string{"小孩", "儿童"}
		// t := time.Now()
		// docs := e.Search(q)
		// st := time.Since(t)
		// t = time.Now()
		// docs = e.FliterResult(docs, filterWord)
		// ft := time.Since(t)
		// for i, doc := range docs {
		// 	if i >= 10 {
		// 		break
		// 	}
		// 	logger.Infoln(doc.Id, doc.Text, (len(doc.Text)+len(doc.Url)+4)/4)
		// 	// fmt.Println(doc.Id, doc.Text)

		// }
		// logger.Infoln("Find docs:", len(docs), "Search time:", st, "Filter time:", ft)
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
	Listen()
	// gen.GenSearchHistory()
	// t := tree.NewTrie()
	// t.LoadData(viper.GetString("db.searchHistory"))
	// logger.Debug(t.RelatedSearch(""))
	logger.Infoln("========================== Process Done ==========================")
}
