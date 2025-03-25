package config

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

const (
  configPath = "/.config/syncer/"
)
var db string

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
  LastPulled string
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

  rows, err := db.Query("SELECT remotename, localpath, status, vendor, lastpulled FROM files")
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  for rows.Next() {
    var remotename string
    var localpath string
    var status Status
    var vendor Vendor
    var lastpulled string

    err = rows.Scan(&remotename, &localpath, &status, &vendor, &lastpulled)
    if err != nil {
      continue
    }

    f := File{}
    f.RemoteName = remotename
    f.LocalPath = localpath
    f.Status = status
    f.Vendor = vendor
    f.LastPulled = lastpulled

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

func Authenticate(vendor Vendor) error {
  // TODO: add check for authentication needs
  switch vendor {
    case GoogleDrive:
      err := authenticateGoogleDrive()
      if err != nil {
        return err
      }
    default:
      fmt.Printf("Vendor %v not supported\n", vendor)
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

func authenticateGoogleDrive() error {
  secret, err := os.ReadFile("secret.json")
  if err != nil {
    return err
  }

  config, err := google.ConfigFromJSON(secret, drive.DriveScope)
  if err != nil {
    return err
  }

  token, err := getTokenFromWeb(config)
  if err != nil {
    return err
  }

  u, err := user.Current()
  if err != nil {
    return err
  }

  p := filepath.Join(u.HomeDir + configPath + "token.json")

  err = saveToken(p, token)
  if err != nil {
    return err
  }

  return nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  exec.Command("xdg-open", authURL).Start()

  code, err := getAuthorizationcodeFromRedirect()
  if err != nil {
    return nil, err
  }

  tok, err := config.Exchange(context.TODO(), code)
  if err != nil {
    return nil, err
  }

  return tok, nil
}

func getAuthorizationcodeFromRedirect() (string, error) {
  listener, err := net.Listen("tcp", "127.0.0.1:3333")
  if err != nil {
    return "", err
  }
  defer listener.Close()

  con, err := listener.Accept()
  if err != nil {
    fmt.Println("Failed to listen for the redirect url")
    return "", err
  }
  defer con.Close()

  tmp := make([]byte, 1024)
  _, err = con.Read(tmp)
  if err != nil {
    return "", err
  }

  // Sanitize data, don´t know if needed
  var data []byte
  for i, v := range tmp  {
    if (v == 0) {
      data = tmp[0:i]
      break
    }
  }

  reader := bufio.NewReader(bytes.NewReader(data))
  req, err := http.ReadRequest(reader)
  if err != nil {
    return "", err
  }
  defer req.Body.Close()
  code := req.URL.Query().Get("code")

  con.Write([]byte("Succes\r\nYou can close this browser window"))

  return code, nil
}

func saveToken(path string, token *oauth2.Token) error {
  file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  if err != nil {
    return err
  }
  defer file.Close() 
  json.NewEncoder(file).Encode(token)

  return nil
}
