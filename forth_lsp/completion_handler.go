package forth_lsp

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Completion(_ context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	line := params.Position.Line
	pos := params.Position.Character

	text, ok := h.documents.Text(params.TextDocument.URI)
	if ok {
		l := strings.Split(text, "\n")[line]
		var w string
		if pos > 0 {
			w = WordAtPoint(l, pos-1)
			w = strings.ToLower(w)
			w = EscapePrefix(w)

		} else {
			w = ""
		}

		log.Println("look up partial word: ", w)

		type resultStruct struct {
			word string
			stack string
			description string
			example string
		}

		var results []resultStruct
		var rows *sql.Rows
		var err error
		if w == "" {
			rows, err = DB.Query("SELECT word, stack, description, example from FORTH ORDER BY word")
		} else {
			q := strings.ToLower(w) + "%"
			rows, err = DB.Query("SELECT word, stack, description, example from FORTH WHERE word LIKE ? ESCAPE '\\' ORDER BY word", q)
		}
		if err != nil {
			log.Printf("DB Query Error result: %v\n", err)
	    	return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var r resultStruct
			if err := rows.Scan(&r.word, &r.stack, &r.description, &r.example); err != nil {
				log.Printf("DB Scan Error result: %v\n", err)
				return nil, err
			}
			results = append(results, r)
		}

		if err := rows.Err(); err != nil {
			log.Printf("DB Scan Err result: %v\n", err)
			return nil, err
		}

		var cil []lsp.CompletionItem
		for _, v := range results {
			ci := lsp.CompletionItem{
				Label: v.word,
				Detail: v.stack,
			}
			cil = append(cil, ci)
		}

		return &lsp.CompletionList{
			Items: cil,
		}, nil
	}

	return nil, nil
}
