package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nib/db"
	"nib/model"

	tea "github.com/charmbracelet/bubbletea"
)

// Version is set via ldflags at build time
var Version = "dev"

func dataHome() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

func main() {
	versionShort := flag.Bool("v", false, "Display version")
	versionLong := flag.Bool("version", false, "Display version")
	flag.Parse()

	if *versionShort || *versionLong {
		fmt.Println(Version)
		return
	}

	dbPath := filepath.Join(dataHome(), "nib", "nib.db")
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
