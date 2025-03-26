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
	"time"

	"syncer/internal/cloud"
	"syncer/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
  ctx := context.Background()

  err := config.Init() 
  if err != nil {
    log.Fatalf("%v", err)
  }

  files, err := config.GetFiles()
  if err != nil {
    log.Fatalf("%v", err)
  }

  for vendor, f := range groupByVendor(files) {
    switch vendor {
      case cloud.GoogleDrive:
        handleGoogleDrive(f, ctx)
    }
  }
}

func handleGoogleDrive(files []cloud.File, ctx context.Context) {
  if len(files) <= 0 {
    return 
  }

  client, err := getClient()
  if err != nil {
    return
  }

  srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
  if err != nil {
    return
  }

  for _, file := range files {
    if file.Vendor != cloud.GoogleDrive {
      continue
    }

    r, err := srv.Files.List().Fields("files(name, id, lastModifyingUser, modifiedTime)").Do()
    if err != nil {
      continue
    }

    // Handle google drive
    var id string
    if len(r.Files) == 0 {
      continue
    } else {
      for _, f := range r.Files {
        modifiedTime, err := time.Parse(time.RFC3339, f.ModifiedTime)
        if err != nil {
          continue
        }
        lastPulled, err := time.Parse(time.RFC3339, file.LastPulled)
        if err != nil {
          continue
        }

        isModifiedSinceLastPull := modifiedTime.Sub(lastPulled) > 0

        if (f.Name == file.RemoteName && isModifiedSinceLastPull) {
          id = f.Id
          break
        }
      }
    }

    if id == "" {
      return
    }

    res, err := srv.Files.Get(id).Download()
    if err != nil {
      log.Fatalf("Unable to download the file")
    }
    defer res.Body.Close()

    saveFile(res, file.LocalPath)
  }
}

func saveFile(response *http.Response, path string) {
  dir := filepath.Dir(path)
  err := os.MkdirAll(dir, 0777)
  if err != nil {
    return 
  }

  destFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

  if err != nil {
    return
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
}

func groupByVendor(files []cloud.File) map[cloud.Vendor][]cloud.File {
  filesPerVendor := make(map[cloud.Vendor][]cloud.File)

  for _, f := range files {
    _, ok := filesPerVendor[f.Vendor]
    if !ok {
      filesPerVendor[f.Vendor] = make([]cloud.File, 0)
    }
    filesPerVendor[f.Vendor] = append(filesPerVendor[f.Vendor], f)
  }

  return filesPerVendor
}

func getClient() (*http.Client, error) {
  secret, err := os.ReadFile("secret.json")
  if err != nil {
    return nil, err
  }

  c, err := google.ConfigFromJSON(secret, drive.DriveScope)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }

  configPath, err := config.GetConfigPath()
  if err != nil {
    return nil, err
  }

  tokFile := filepath.Join(configPath, "google_token.json")

  file, err := os.Open(tokFile)
  if err != nil {
    return nil, err
  }
  defer file.Close()
  tok := &oauth2.Token{}
  _ = json.NewDecoder(file).Decode(tok)

  return c.Client(context.Background(), tok), nil
}
