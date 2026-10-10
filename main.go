package main

import (
	"context"
	"log"
	"os"

	"github.com/owenrumney/go-lsp/server"
	lsp "github.com/wolfmanjm/wjm_forth_lsp/forth_lsp"
	forthdoc "github.com/wolfmanjm/wjm_forth_lsp/forth_parser"
)

func main() {
	// if we have any args we generate forth docs otherwise we run the LSP
	if len(os.Args) > 1 {
		forthdoc.GenerateDocs()
		return
	}

	// LSP
	h := lsp.New()
	srv := server.NewServer(h)
	if err := srv.Run(context.Background(), server.RunStdio()); err != nil {
		log.Fatal(err)
	}
}
