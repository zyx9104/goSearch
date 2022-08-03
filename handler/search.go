package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/z-y-x233/goSearch/pkg/model"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"github.com/z-y-x233/goSearch/pkg/tree"
)

func Search(req *model.SearchRequest) *model.SearchResponse {
	start := time.Now()

	//获取分词
	cutWords := tools.WordCut(req.Query)
	//获取文档
	slices := e.Query(req.Query)
	docs := e.GetDocs(slices)
	docs = e.FliterResult(docs, req.FilterWord)
	respDocs := make([]model.ResponseDoc, 0, len(docs))
	for i, doc := range docs {
		for _, word := range cutWords {
			doc.Text = strings.ReplaceAll(doc.Text, word, req.Highlight.PreTag+word+req.Highlight.PostTag)
		}
		respDocs = append(respDocs, model.ResponseDoc{Id: doc.Id, Text: doc.Text, Url: doc.Url, Score: slices[i].Score})
	}
	//相关搜索
	qs := tree.FindRelated(req.Query, 10)
	related := []string{}
	for _, q := range qs {
		related = append(related, q.Q)
	}
	total := len(respDocs)
	pageCount := (total + req.Limit - 1) / req.Limit
	//分页
	begin := (req.Page - 1) * req.Limit
	end := begin + req.Limit
	if begin > len(respDocs) {
		begin = len(respDocs)
	}
	if end > len(respDocs) {
		end = len(respDocs)
	}
	respDocs = respDocs[begin:end]

	//添加查询
	tree.AddQuery(req.Query)

	result := &model.SearchResponse{
		Total:     total,
		PageCount: pageCount,
		Page:      req.Page,
		Limit:     req.Limit,
		Documents: respDocs,
		Related:   related,
		Words:     cutWords,
	}
	result.Time = float64(time.Since(start) / time.Millisecond)
	return result
}

func Put(c *gin.Context) {

}

func Get(c *gin.Context) {

}

func Related(c *gin.Context) {

}
