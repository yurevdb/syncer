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

  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return nil, err
  }
  defer db.Close()

  rows, err := db.Query("SELECT remotename, localpath, status FROM files")
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  for rows.Next() {
    var remotename string
    var localpath string
    var status Status
    err = rows.Scan(&remotename, &localpath, &status)
    if err != nil {
      continue
    }
    f := File{}
    f.RemoteName = remotename
    f.LocalPath = localpath
    f.Status = status
    files = append(files, f)
  }

  return files, nil
}

func AddFile(file File) error {
  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return err
  }
  defer db.Close()

  stmt, err := db.Prepare("INSERT INTO files(remotename, localpath, status) VALUES (?, ?, ?)")
  if err != nil {
    return err
  }
  defer stmt.Close()
  _, err = stmt.Exec(file.RemoteName, file.LocalPath, Error)
  if err != nil {
    return err
  }

  return nil
}

func RemoveFile(remoteName string) error {
  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return err
  }
  defer db.Close()

  stmt, err := db.Prepare("DELETE FROM files WHERE remotename = ?")
  if err != nil {
    return err
  }
  defer stmt.Close()
  _, err = stmt.Exec(remoteName)
  if err != nil {
    return err
  }

  return nil
}

func createTables(db *sql.DB) error {
  query := `CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            remotename TEXT UNIQUE,
            localpath TEXT UNIQUE,
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
