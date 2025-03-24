package config

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
  db = "~/.config/syncer/syncer.db"
  dbDir = "~/.config/syncer/"
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

type Vendor int
const (
  GoogleDrive Vendor = iota
)
var vendorString = map[Vendor]string{
  GoogleDrive: "Google Drive",
}
func (vendor Vendor) String() string {
  return vendorString[vendor]
}

type File = struct {
  Status Status
  Vendor Vendor
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

  rows, err := db.Query("SELECT remotename, localpath, status, vendor FROM files")
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  for rows.Next() {
    var remotename string
    var localpath string
    var status Status
    var vendor Vendor
    err = rows.Scan(&remotename, &localpath, &status, &vendor)
    if err != nil {
      continue
    }
    f := File{}
    f.RemoteName = remotename
    f.LocalPath = localpath
    f.Status = status
    f.Vendor = vendor
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

  stmt, err := db.Prepare("INSERT INTO files(remotename, localpath, status, vendor) VALUES (?, ?, ?, ?)")
  if err != nil {
    return err
  }
  defer stmt.Close()
  _, err = stmt.Exec(file.RemoteName, file.LocalPath, Error, file.Vendor)
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
            status INTEGER,
            vendor INTEGER)`

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
