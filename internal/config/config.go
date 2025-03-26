package config

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

  "syncer/internal/cloud"

	_ "github.com/mattn/go-sqlite3"
)

const (
  configPath = "/.config/syncer/"
)
var db string


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

func GetFiles() ([]cloud.File, error) {
  files := make([]cloud.File, 0)

  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return nil, err
  }
  defer db.Close()

  rows, err := db.Query("SELECT remotename, localpath, status, vendor, lastpulled FROM files")
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  for rows.Next() {
    var remotename string
    var localpath string
    var status cloud.Status
    var vendor cloud.Vendor
    var lastpulled string

    err = rows.Scan(&remotename, &localpath, &status, &vendor, &lastpulled)
    if err != nil {
      continue
    }

    f := cloud.File{}
    f.RemoteName = remotename
    f.LocalPath = localpath
    f.Status = status
    f.Vendor = vendor
    f.LastPulled = lastpulled

    files = append(files, f)
  }

  return files, nil
}

func AddFile(file cloud.File) error {
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

  _, err = stmt.Exec(file.RemoteName, file.LocalPath, cloud.Error, file.Vendor)
  if err != nil {
    fmt.Printf("Error: %v\n", err)
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

func GetConfigPath() (string, error) {
  u, err := user.Current()
  if err != nil {
    return "", err
  }

  p, _ := filepath.Abs(filepath.Join(u.HomeDir, configPath))

  return p, nil
}

func createTables(db *sql.DB) error {
  query := `CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            remotename TEXT NOT NULL,
            localpath TEXT NOT NULL UNIQUE,
            status INTEGER NOT NULL,
            vendor INTEGER NOT NULL,
            lastpulled DATETIME DEFAULT(DATETIME('1900-01-01 00:00:00')) NOT NULL,
            UNIQUE(remotename, vendor))`

  _, err := db.Exec(query)
  if err != nil {
    return err
  }

  return nil
}

func ensureDatabasePathExists() error {
  u, err := user.Current()
  if err != nil {
    return err
  }

  p, _ := filepath.Abs(u.HomeDir +  configPath)
  db = p + "/syncer.db"

  err = os.MkdirAll(p, 0777)
  if err != nil {
    return err
  }

  return nil
}

