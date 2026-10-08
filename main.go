package main

import (
	"context"
	"log"

	"github.com/owenrumney/go-lsp/server"
	lsp "github.com/wolfmanjm/wjm_forth_lsp/forth_lsp"
)

func main() {
	h := lsp.New()
	srv := server.NewServer(h)
	if err := srv.Run(context.Background(), server.RunStdio()); err != nil {
		log.Fatal(err)
	}
}
