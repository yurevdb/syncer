package main

import (
	"log"

	"syncer/internal/cloud"
	"syncer/internal/config"
)

func main() {
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
        handleGoogleDrive(f)
    }
  }
}

func handleGoogleDrive(files []cloud.File) {
  for _, f := range files {
    cloud.GoogleDrive.Repository().Pull(&f)
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
