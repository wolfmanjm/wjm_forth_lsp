package forth_lsp

import (
	"context"

	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Completion(_ context.Context, _ *lsp.CompletionParams) (*lsp.CompletionList, error) {
	return &lsp.CompletionList{
		Items: []lsp.CompletionItem{
			{Label: "example"},
		},
	}, nil
}
