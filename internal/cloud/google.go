package cloud

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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

  //var id string
  for _, f := range r.Items {
    // Checks if remote file was updated since last pull

    //modifiedTime, err := time.Parse(time.RFC3339, f.ModifiedDate)
    //if err != nil {
    //  continue
    //}
    //lastPulled, err := time.Parse(time.RFC3339, file.LastPulled)
    //if err != nil {
    //  continue
    //}

    //isModifiedSinceLastPull := modifiedTime.Sub(lastPulled) > 0

    //if (f.OriginalFilename == file.RemoteName && isModifiedSinceLastPull) {
    //  id = f.Id
    //  break
    //}

    //if id == "" {
    //  return nil
    //}

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
    
    // TODO: update file, lastpulled & status
  }
  return nil
}

func (g Google) PullAll(files *[]File) error {

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
