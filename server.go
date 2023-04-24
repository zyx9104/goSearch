package main

import (
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"

	"github.com/z-y-x233/goSearch/api"
	"github.com/z-y-x233/goSearch/handler"
	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/log"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"github.com/z-y-x233/goSearch/pkg/tree"
)

var (
	g *gin.Engine
)

func init() {
	engine.Init()
	tools.Init()
	tree.Init()
	handler.Init()
	g = gin.Default()
	g.Use(api.Cors())
	api.InitRouter(g)

}

func close() {
	handler.Close()
	tree.Close()
}

func main() {
	defer close()
	log.Infoln("========================== Process Start ==========================")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	gin.SetMode(gin.ReleaseMode)
	go g.Run(":8080")
	<-c
	log.Infoln("========================== Process Done ==========================")
	log.Infoln("========================== Save Data ==========================")

}
