package handler

import (
	"github.com/z-y-x233/goSearch/pkg/engine"
)

var (
	e *engine.Engine
)

func Init() {
	e = engine.DefaultEngine()
}

func Close() {
	e.Close()
}
