package cloud

import (
	"os/user"
	"path/filepath"
)

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
func (vendor Vendor) Repository() Repository {
  u, _ := user.Current()

  switch vendor {
    case GoogleDrive:
      p := filepath.Join(u.HomeDir, ".config/syncer", "google_token.json")
      return Google{
        TokenPath: p,
      }
    default:
      return nil
  }
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

type Repository interface {
  Pull(file *File) error
  Push(file *File) error
  List() ([]string, error)
  Authenticate() error
}
