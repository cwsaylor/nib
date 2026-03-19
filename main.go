package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nib/db"
	"nib/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adrg/xdg"
)

func main() {
	dbPath := filepath.Join(xdg.DataHome, "nib", "nib.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		os.Exit(1)
	}

	store, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	purged, _ := store.PurgeExpiredTrash()
	if purged > 0 {
		log.Printf("Purged %d expired notes from trash", purged)
	}

	p := tea.NewProgram(
		model.New(store),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
