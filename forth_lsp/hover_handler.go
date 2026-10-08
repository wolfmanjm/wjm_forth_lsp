package forth_lsp

import (
	"context"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Hover(_ context.Context, _ *lsp.HoverParams) (*lsp.Hover, error) {
	return &lsp.Hover{
		Contents: lsp.NewHoverContents(lsp.Markdown, "**mylang** hover"),
	}, nil
}
