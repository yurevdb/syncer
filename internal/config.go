package internal

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

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

func GetFiles() ([]File, error) {
  files := make([]File, 0)

  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return nil, err
  }
  defer db.Close()

  rows, err := db.Query(`SELECT   id, 
                                  remotename, 
                                  localpath, 
                                  status, 
                                  vendor, 
                                  lastpulled 
                         FROM     files`)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  for rows.Next() {
    var id int
    var remotename string
    var localpath string
    var status Status
    var vendor Vendor
    var lastpulled string

    err = rows.Scan(&id, &remotename, &localpath, &status, &vendor, &lastpulled)
    if err != nil {
      continue
    }

    f := File{
      Id: id,
      RemoteName: remotename,
      LocalPath: localpath,
      Status: status,
      Vendor: vendor,
      LastPulled: lastpulled,
    }

    files = append(files, f)
  }

  return files, nil
}

func UpdateFile(file File) error {
  db, err := sql.Open("sqlite3", db)
  if err != nil {
    return err
  }
  defer db.Close()

  stmt, err := db.Prepare("UPDATE files SET status = ?, lastpulled = CURRENT_TIMESTAMP WHERE id = ?")
  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec(file.Status, file.Id)
  if err != nil {
    fmt.Printf("Error: %v\n", err)
    return err
  }

  return nil
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

func createTables(db *sql.DB) error {
  query := `CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            remotename TEXT NOT NULL,
            localpath TEXT NOT NULL UNIQUE,
            status INTEGER NOT NULL,
            vendor INTEGER NOT NULL,
            lastpulled DATETIME DEFAULT(CURRENT_TIMESTAMP) NOT NULL,
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

