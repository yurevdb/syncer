package config

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
  db = "/home/yure/.config/syncer/syncer.db"
  dbDir = "/home/yure/.config/syncer/"
)

type Status int

const (
  Error Status = iota
  Synced
)

var statusString = map[Status]string{
  Error: "Error",
  Synced: "Synced",
}

func (status Status) String() string{
  return statusString[status]
}

type File = struct {
  Status Status
  RemoteName string
  LocalPath string
}

func Init() error {
  ensureDatabasePathExists()

  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return err
  }
  defer db.Close()

  err = createTables(db)
  if err != nil {
    return err
  }

  return nil
}

func GetFiles() ([]File, error) {
  files := make([]File, 0)

  f := File{}
  f.RemoteName = "Test"
  f.LocalPath = "~/google-drive/Test"
  f.Status = Synced
  files = append(files, f)

  ff := File{}
  ff.RemoteName = "MainPasswords.kdbx"
  ff.LocalPath = "~/google-drive/MainPasswords.kdbx"
  ff.Status = Error
  files = append(files, ff)

  return files, nil
}

func createTables(db *sql.DB) error {
  query := `CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            remotename TEXT,
            localpath TEXT,
            status INTEGER)`

  _, err := db.Exec(query)
  if err != nil {
    return err
  }

  return nil
}

func ensureDatabasePathExists() error {
  p, _ := filepath.Abs(dbDir)
  err := os.MkdirAll(p, os.ModeDir)
  if err != nil {
    return err
  }

  return nil
}
