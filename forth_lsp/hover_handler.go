package forth_lsp

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/owenrumney/go-lsp/lsp"

	"database/sql"

	_ "modernc.org/sqlite"
	"modernc.org/sqlite/vfs"
)

var DB *sql.DB

//go:embed data/mecrisp_stellaris.db
var embeddedFS embed.FS

func OpenDatabase() (error) {
	// Register the embed.FS to create a new Virtual File System
	vfsName, _, err := vfs.New(embeddedFS)
	if err != nil {
		return fmt.Errorf("Failed to initialize VFS: %v", err)
	}

	// Connect to the file inside the VFS by appending '?vfs=' to the DSN
	// Note: Use 'mode=ro' to explicitly tell SQLite this is a read-only instance
	dsn := fmt.Sprintf("file:data/mecrisp_stellaris.db?vfs=%s&mode=ro", vfsName)

	// Open the database using the dsn above
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("Unable to open database file - %w", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Unable to ping database file: %w", err)
	}

	log.Println("Database opened")

	DB= db

	return nil
}

func CloseDatabase() {
	DB.Close()
	log.Println("Database closed")
	DB = nil
}

func EscapePrefix(p string) string {
    p = strings.ReplaceAll(p, "\\", "\\\\")
    p = strings.ReplaceAll(p, "%", "\\%")
    p = strings.ReplaceAll(p, "_", "\\_")
    return p
}

func (h *Handler) Hover(_ context.Context, params *lsp.HoverParams) (*lsp.Hover, error) {

	// fmt.Fprintln(os.Stderr, params)

	line:= params.Position.Line
	pos := params.Position.Character

	text, ok := h.documents.Text(params.TextDocument.URI)
	if ok {
		l := strings.Split(text, "\n")[line]
		w := WordAtPoint(l, pos)
		if w == "" { return nil, nil }
		w = strings.ToLower(w)
		w = EscapePrefix(w)
		// log.Println("look up word: ", w)

		var (
			word string
			stack string
			description string
			example string
		)
	    row := DB.QueryRow("SELECT word, stack, description, example from FORTH WHERE word LIKE ? ESCAPE '\\'", w)
	    err := row.Scan(&word, &stack, &description, &example);
	    if err == nil {
    		// log.Printf("DB result: %s, %s, %s\n", word, stack, description)
			r := fmt.Sprintf("*%s*  `%s`  \n%s  \n", word, stack, description)

			return &lsp.Hover{
				Contents: lsp.NewHoverContents(lsp.Markdown, r),
			}, nil
		} else {
	    	// log.Printf("DB Error result: %v\n", err)
	    	return nil, nil
		}
	}
	return nil, nil
}
