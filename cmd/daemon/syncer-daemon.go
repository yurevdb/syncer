package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
  ctx := context.Background()
  secret, err := os.ReadFile("secret.json")
  if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
  }

  config, err := google.ConfigFromJSON(secret, drive.DriveScope)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }
  client := getClient(config)

  srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
  if err != nil {
    log.Fatalf("Unable to retrieve Drive client: %v", err)
  }

  // TODO: get this from config 
  filename := "MainPasswords.kdbx"

  r, err := srv.Files.List().Fields("nextPageToken, files(id, name)").Do()
  if err != nil {
    log.Fatalf("Unable to retrieve files: %v", err)
  }

  // Handle google drive
  var id string
  if len(r.Files) == 0 {
    fmt.Println("No files found.")
    os.Exit(1)
  } else {
    for _, file := range r.Files {
      if (file.Name == filename) {
        // TODO: use this to keep track of download needs
        //file.ModifiedTime
        id = file.Id
        fmt.Printf("Found %v\n", file.Name)
        break
      }
    }
  }

  res, err := srv.Files.Get(id).Download()
  if err != nil {
    log.Fatalf("Unable to download the file")
    fmt.Printf("%v\n", err)
  }
  defer res.Body.Close()

  destFile, err := os.OpenFile("./test/MainPasswords.kdbx", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  buffer := make([]byte, 32 * 1024)
  pos := 0
  for {
    read, err := res.Body.Read(buffer)

    if read > 0 {
      written, err := destFile.WriteAt(buffer, int64(pos))
      if err != nil {
        log.Fatalf("Unable to write downloaded file to local file")
        break
      }
      pos += written
    }

    if err != nil {
      if err == io.EOF {
        break
      }

      log.Fatalf("Problem with downloading file")
      break
    }
  }
}

func getClient(config *oauth2.Config) *http.Client {
  tokFile := "token.json"
  tok, err := tokenFromFile(tokFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(tokFile, tok)
  }
  return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  exec.Command("xdg-open", authURL).Start()

  code, err := getAuthorizationcodeFromRedirect()
  if err != nil {
    log.Fatalf("Unable to get authorization code from redirect")
  }

  tok, err := config.Exchange(context.TODO(), code)
  if err != nil {
    log.Fatalf("Uable to retrieve token from web %v", err)
  }

  return tok
}

func getAuthorizationcodeFromRedirect() (string, error) {
  listener, err := net.Listen("tcp", "127.0.0.1:3333")
  if err != nil {
    return "", err
  }
  defer listener.Close()

  con, err := listener.Accept()
  if err != nil {
    log.Fatalf("Failed to listen for the redirect url")
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

func tokenFromFile(path string) (*oauth2.Token, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()
  tok := &oauth2.Token{}
  err = json.NewDecoder(file).Decode(tok)
  return tok, err
}

func saveToken(path string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", path)
  file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  if err != nil {
    log.Fatalf("Unable to cache oauth toke: %v", err)
  }
  defer file.Close() 
  json.NewEncoder(file).Encode(token)
}
