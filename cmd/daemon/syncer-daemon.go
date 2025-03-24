package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
  "syncer/internal/config"
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
  client, err := getClient(config)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }

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

func getClient(c *oauth2.Config) (*http.Client, error) {
  configPath, err := config.GetConfigPath()
  if err != nil {
    return nil, err
  }

  tokFile := filepath.Join(configPath, "token.json")

  file, err := os.Open(tokFile)
  if err != nil {
    return nil, err
  }
  defer file.Close()
  tok := &oauth2.Token{}
  _ = json.NewDecoder(file).Decode(tok)

  return c.Client(context.Background(), tok), nil
}
