package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

type Google struct {
  TokenPath string
}

func (g Google) Pull(file *File) error {
  client, err := getClient(&g)
  if err != nil {
    return err
  }

  srv, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
  if err != nil {
    return err
  }

  r, err := srv.Files.List().Do()
  if err != nil {
    return err
  }

  for _, f := range r.Items {
    if (f.OriginalFilename != file.RemoteName) {
      continue
    }

    res, err := srv.Files.Get(f.Id).Download()
    if err != nil {
      return err
    }
    defer res.Body.Close()

    err = saveFile(res, file.LocalPath)
    if err != nil {
      return err
    }
    
    file.Status = Synced
    UpdateFile(*file)
  }
  return nil
}

func (g Google) PullAll(files []File) error {
  client, err := getClient(&g)
  if err != nil {
    return err
  }

  srv, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
  if err != nil {
    return err
  }

  r, err := srv.Files.List().Do()
  if err != nil {
    return err
  }

  for _, lf := range files {
    for _, rf := range r.Items {

      if (rf.OriginalFilename != lf.RemoteName) {
        continue
      }

      res, err := srv.Files.Get(rf.Id).Download()
      if err != nil {
        return err
      }
      defer res.Body.Close()

      err = saveFile(res, lf.LocalPath)
      if err != nil {
        return err
      }
      
      lf.Status = Synced
      UpdateFile(lf)
    }
  }

  return nil
}

func saveFile(response *http.Response, path string) error {
  dir := filepath.Dir(path)
  err := os.MkdirAll(dir, 0777)
  if err != nil {
    return err
  }

  destFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

  if err != nil {
    return err
  }

  buffer := make([]byte, 32 * 1024)
  pos := 0
  for {
    read, err := response.Body.Read(buffer)

    if read > 0 {
      written, err := destFile.WriteAt(buffer, int64(pos))
      if err != nil {
        break
      }
      pos += written
    }

    if err != nil {
      if err == io.EOF {
        break
      }
      break
    }
  }

  return nil
}

func (g Google) Push(file *File) error {
  client, err := getClient(&g)
  if err != nil {
    return err
  }

  srv, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
  if err != nil {
    return err
  }

  r, err := srv.Files.List().Do()
  if err != nil {
    return err
  }

  var rf_id string = ""
  for _, rf := range r.Items {
    if rf.OriginalFilename != file.RemoteName {
      continue
    }
    rf_id = rf.Id
  }

  if rf_id == "" {
    // Insert file
    lf, err := os.Open(file.LocalPath)
    if err != nil {
      return err
    }
    defer lf.Close()

    f := &drive.File{
      Title: file.RemoteName,
    }

    _, err = srv.Files.Insert(f).Media(lf).Do()
    if err != nil {
      return err
    }
  } else {
    // Update file
    // Insert file
    lf, err := os.Open(file.LocalPath)
    if err != nil {
      return err
    }
    defer lf.Close()

    rf, err := srv.Files.Get(rf_id).Do()
    if err != nil {
      return err
    }

    _, err = srv.Files.Update(rf.Id, rf).Media(lf).Do()
    if err != nil {
      return err
    }

    file.Status = Synced
    err = UpdateFile(*file)
    if err != nil {
      fmt.Printf("Update File: %v\n", err)
      return err
    }
  }

  return nil
}

func (g Google) PushAll(files []File) error {
  errors := make([]string, 0)

  for _, lf := range files {
    err := g.Push(&lf)
    if err != nil {
      errors = append(errors, err.Error())
    }
  }

  // Hanlde Errors
  if len(errors) > 0 {
    return fmt.Errorf(strings.Join(errors, "\n"))
  }

  return nil
}

func (g Google) List() ([]string, error) {
  client, err := getClient(&g)
  if err != nil {
    return nil, err
  }

  srv, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
  if err != nil {
    return nil, err
  }

  // TODO: fields don't work
  r, err := srv.Files.List().Do()
  if err != nil {
    return nil, err
  }

  remoteFiles := make([]string, 0)
  for _, f := range r.Items {
    if f.OriginalFilename == "" {
      continue
    }
    remoteFiles = append(remoteFiles, " - " + f.OriginalFilename)
  }

  return remoteFiles, nil
}

func getClient(g *Google) (*http.Client, error) {
  secret, err := os.ReadFile("secret.json")
  if err != nil {
    return nil, err
  }

  c, err := google.ConfigFromJSON(secret, drive.DriveScope)
  if err != nil {
    return nil, err
  }

  file, err := os.Open(g.TokenPath)
  if err != nil {
    return nil, err
  }
  defer file.Close()
  tok := &oauth2.Token{}
  _ = json.NewDecoder(file).Decode(tok)

  return c.Client(context.Background(), tok), nil
}

func (g Google) Authenticate() error {
  _, err := os.Open(g.TokenPath)
  if err == nil {
    return nil
  }

  // TODO: find a better way to handle the secret
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

  err = saveToken(g.TokenPath, token)
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

func (g Google) IsAuthenticated() bool {
  _, err := os.Open(g.TokenPath)
  if err != nil {
    return false
  }

  return true
}
