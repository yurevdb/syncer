package main

import (
	"log"
	"time"

	"syncer/internal"
)

func main() {
  err := internal.Init() 
  if err != nil {
    log.Fatalf("%v", err)
  }

  for {
    time.Sleep(time.Second * 60 * 15)

    files, err := internal.GetFiles()
    if err != nil {
      continue
    }

    for vendor, f := range groupByVendor(files) {
      switch vendor {
      case internal.GoogleDrive:
        handleGoogleDrive(f)
      }
    }
  }
}

func handleGoogleDrive(files []internal.File) {
  for _, f := range files {
    internal.GoogleDrive.Repository().Pull(&f)
    internal.GoogleDrive.Repository().Push(&f)
  }
}

func groupByVendor(files []internal.File) map[internal.Vendor][]internal.File {
  filesPerVendor := make(map[internal.Vendor][]internal.File)

  for _, f := range files {
    _, ok := filesPerVendor[f.Vendor]
    if !ok {
      filesPerVendor[f.Vendor] = make([]internal.File, 0)
    }
    filesPerVendor[f.Vendor] = append(filesPerVendor[f.Vendor], f)
  }

  return filesPerVendor
}
