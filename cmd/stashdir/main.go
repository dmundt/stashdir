package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmundt/stashdir/internal/store"

	"github.com/atotto/clipboard"
)

func usage() {
	fmt.Println("stashdir - bookmark directories")
	fmt.Println("Usage:")
	fmt.Println("  stashdir add [PATH]       Add current or specified path")
	fmt.Println("  stashdir copy [IDX]       Interactive or by 1-based index; copies chosen path to clipboard")
	fmt.Println("  stashdir list             List saved paths")
	fmt.Println("  stashdir path             Show absolute path to the database file")
	fmt.Println("  stashdir remove <ARG>     Remove by index (1-based) or by exact path")
	fmt.Println("  stashdir select [IDX]     Interactive or by 1-based index; prints chosen path")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	cmd := os.Args[1]
	db, err := store.Open()
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	switch cmd {
	case "add":
		var p string
		if len(os.Args) >= 3 {
			p = os.Args[2]
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("getwd: %v", err)
			}
			p = cwd
		}
		p, err = filepath.Abs(p)
		if err != nil {
			log.Fatalf("abs path: %v", err)
		}
		if err := db.Add(p); err != nil {
			log.Fatalf("add: %v", err)
		}
	case "list":
		items := db.List()
		for i, it := range items {
			fmt.Printf("%d\t%s\n", i+1, it)
		}
	case "select":
		if len(os.Args) >= 3 {
			idx, err := parseIndex(os.Args[2])
			if err != nil {
				log.Fatalf("invalid index: %v", err)
			}
			items := db.List()
			if idx <= 0 || idx > len(items) {
				log.Fatalf("index out of range (1..%d)", len(items))
			}
			fmt.Print(items[idx-1])
			return
		}
		choice, err := db.SelectInteractive()
		if err != nil {
			log.Fatalf("select: %v", err)
		}
		if choice != "" {
			fmt.Print(choice)
		}
	case "copy":
		var choice string
		if len(os.Args) >= 3 {
			idx, err := parseIndex(os.Args[2])
			if err != nil {
				log.Fatalf("invalid index: %v", err)
			}
			items := db.List()
			if idx <= 0 || idx > len(items) {
				log.Fatalf("index out of range (1..%d)", len(items))
			}
			choice = items[idx-1]
		} else {
			var err error
			choice, err = db.SelectInteractive()
			if err != nil {
				log.Fatalf("select: %v", err)
			}
			if choice == "" {
				return
			}
		}
		if err := clipboard.WriteAll(choice); err != nil {
			log.Fatalf("copy to clipboard: %v", err)
		}
		fmt.Println("Copied to clipboard:", choice)
	case "path":
		// Print the absolute path to the database file so users can inspect or back it up
		p := db.Path
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		fmt.Print(p)
	case "remove":
		if len(os.Args) < 3 {
			log.Fatal("remove requires an index or path")
		}
		arg := os.Args[2]
		// Try index first
		if idx, err := parseIndex(arg); err == nil {
			if err := db.RemoveIndex(idx - 1); err != nil {
				log.Fatalf("remove index: %v", err)
			}
			return
		}
		// Fall back to exact path
		if err := db.RemovePath(arg); err != nil {
			log.Fatalf("remove path: %v", err)
		}
	default:
		usage()
	}
}

func parseIndex(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid index: %s", s)
	}
	return n, nil
}
