package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/tnm-yes-anyidea/auD-limit/internal/db"
	"github.com/tnm-yes-anyidea/auD-limit/internal/scanner"
	"github.com/tnm-yes-anyidea/auD-limit/internal/ui"
)

func main() {
	scanPath := flag.String("scan", "", "Path to scan for audio files (optional)")
	flag.Parse()

	// initialize DB (file in current dir)
	dbpath := ".audlimiter.db"
	dbConn, err := db.Open(dbpath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer dbConn.Close()

	// optional scan
	if *scanPath != "" {
		abs, err := filepath.Abs(*scanPath)
		if err != nil {
			log.Fatalf("invalid scan path: %v", err)
		}
		log.Printf("Scanning %s ...", abs)
		if err := scanner.ScanAndStore(dbConn, abs); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
	}

	tracks, err := dbConn.ListTracks()
	if err != nil {
		log.Fatalf("list tracks: %v", err)
	}

	p, err := ui.NewProgram(tracks)
	if err != nil {
		log.Fatalf("tui init: %v", err)
	}

	if err := p.Start(); err != nil {
		log.Printf("tui error: %v", err)
		os.Exit(1)
	}
}