package main

import (
	"context"
	"log"

	"github.com/owenrumney/go-lsp/server"
	"github.com/wolfmanjm/wjm_forth_lpc/forth_lsp"
)

func main() {
	h := forth_lsp.New()
	srv := server.NewServer(h)
	if err := srv.Run(context.Background(), server.RunStdio()); err != nil {
		log.Fatal(err)
	}
}
