package main

import (
	handler "pzfunc"

	"github.com/scaleway/serverless-functions-go/local"
)

func main() {
	local.ServeHandler(handler.Handler, local.WithPort(8080))
}
