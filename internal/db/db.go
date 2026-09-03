package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Track represents a simple music item
type Track struct {
	ID       int64
	Path     string
	Title    string
	Artist   string
	Album    string
	Duration float64
}

type DB struct{
	x *sql.DB
}

func Open(path string) (*DB, error) {
	x, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := x.Exec(`CREATE TABLE IF NOT EXISTS tracks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE,
		title TEXT,
		artist TEXT,
		album TEXT,
		duration REAL
	)`); err != nil {
		x.Close()
		return nil, err
	}
	return &DB{x: x}, nil
}

func (d *DB) Close() error { return d.x.Close() }

func (d *DB) UpsertTrack(t Track) error {
	stmt := `INSERT INTO tracks (path,title,artist,album,duration) VALUES (?,?,?,?,?)
		ON CONFLICT(path) DO UPDATE SET title=excluded.title, artist=excluded.artist, album=excluded.album, duration=excluded.duration;`
	_, err := d.x.Exec(stmt, t.Path, t.Title, t.Artist, t.Album, t.Duration)
	return err
}

func (d *DB) ListTracks() ([]Track, error) {
	rows, err := d.x.Query("SELECT id, path, title, artist, album, duration FROM tracks ORDER BY artist, album, title")
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Track
	for rows.Next() {
		var t Track
		if err := rows.Scan(&t.ID, &t.Path, &t.Title, &t.Artist, &t.Album, &t.Duration); err != nil { return nil, err }
		out = append(out, t)
	}
	return out, nil
}

func (d *DB) DeleteAll() error {
	_, err := d.x.Exec("DELETE FROM tracks")
	return err
}

func (d *DB) Count() (int64, error) {
	var n int64
	row := d.x.QueryRow("SELECT COUNT(1) FROM tracks")
	if err := row.Scan(&n); err != nil { return 0, err }
	return n, nil
}