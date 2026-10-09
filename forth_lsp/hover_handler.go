package forth_lsp

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
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
		log.Fatalf("Failed to initialize VFS: %v", err)
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

func (h *Handler) Hover(_ context.Context, params *lsp.HoverParams) (*lsp.Hover, error) {

	fmt.Fprintln(os.Stderr, params)

	line:= params.Position.Line
	pos := params.Position.Character

	text, ok := h.documents.Text(params.TextDocument.URI)
	if ok {
		line := strings.Split(text, "\n")[line]
		w := WordAtPoint(line, pos)
		if w == "" { return nil, nil }

		// log.Println("look up word: ", w)

		var (
			id int
			word string
			stack string
			description string
			example string
		)
	    row := DB.QueryRow("SELECT id, word, stack, description, example from FORTH WHERE word = ?", w)
	    err := row.Scan(&id, &word, &stack, &description, &example);
	    if err == nil {
    		// log.Printf("DB result: %s, %s, %s\n", word, stack, description)
			r := fmt.Sprintf("*%s*  \n**stack**: %s  \n%s  \n", word, stack, description)

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
