package handler

import (
	"github.com/gorilla/mux"
)

type Handler interface {
	FillHandlers(router *mux.Router)
	Shutdown()
}
