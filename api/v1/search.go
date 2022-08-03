package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z-y-x233/goSearch/handler"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/model"
)

// curl -H "Content-Type: application/json" -X POST -d '{"query": "123"}' localhost:8080/api/v1/search

func Search(c *gin.Context) {
	var request = &model.SearchRequest{}
	err := c.BindJSON(request)
	if err != nil {
		logger.Debug(err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err})
		return
	}
	logger.Debug(request)
	request.GetAndSetDefault()
	result := handler.Search(request)
	c.JSON(http.StatusOK, result)
}

func Put(c *gin.Context) {

}

func Get(c *gin.Context) {

}

func Related(c *gin.Context) {

}
