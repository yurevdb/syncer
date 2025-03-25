package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"syncer/internal/config"
)

func main() {
  if runtime.GOOS == "windows" {
    fmt.Println("Windows is not yet implemented")
    os.Exit(1)
  }

  err := config.Init()
  if err != nil {
    fmt.Printf("Unable to initialize the configuration\n")
    fmt.Printf("%v\n", err)
    os.Exit(1)
  }

  if len(os.Args) == 1 {
    printGeneralHelp()
    os.Exit(0)
  }

  handleCommand(os.Args[1], os.Args[2:]...)
}

func handleCommand(command string, args ...string) {
  switch strings.ToLower(command) {
    case "pull":
    case "browse":
    case "auth":
      handleAuth()
    case "status":
      handleStatus()
    case "start":
    case "stop":
    case "ls":
      handleLs()
    case "add":
      handleAdd(args)
    case "rm":
      handleRm(args)
    case "help":
      if len(args) > 0 {
        printCommandHelp(args[0])
      } else {
        printGeneralHelp()
      }
    default:
      printGeneralHelp()
  }
}

func handleAuth() {
  // TODO: handle vendoring
  err := config.Authenticate(config.GoogleDrive)
  if err != nil {
    fmt.Printf("Unable to authenticate for %v\n%v\n", config.GoogleDrive, err)
  }
}

func handleStatus() {
  ids, err := findPid("syncer-daemon")
  if err != nil {
    fmt.Println("Unable to check if the daemon is running")
  }
  if len(ids) > 0 {
    fmt.Println("Syncer is \033[0;32mrunning\033[0;37m")
  } else {
    fmt.Println("Syncer has \033[0;31mstopped\033[0;37m")
  }

  files, err := config.GetFiles()
  if err != nil {
    fmt.Println("Unable to list files")
  }

  if len(files) > 0 {
    fmt.Println()
  }

  for _, f := range files{
    switch f.Status {
      case config.Error:
        fmt.Printf("%v (%v): \033[0;31m%v\033[0;37m\n", f.RemoteName, f.Vendor, f.Status)
      case config.Synced:
        fmt.Printf("%v (%v): \033[0;32m%v\033[0;37m\n", f.RemoteName, f.Vendor, f.Status)
      default:
        fmt.Printf("%v: Unknown\n", f.RemoteName)
    }
  }

  fmt.Println()
}

func findPid(name string) ([]int, error) {
  ids := make([]int, 0)

  procs, err := os.ReadDir("/proc")
  if err != nil {
    return nil, err
  }

  for _, proc := range procs {
    pid, err := strconv.Atoi(proc.Name())
    if err != nil {
      continue
    }

    binPath := filepath.Join("/proc", proc.Name(), "exe")

    link, err := os.Readlink(binPath)
    if err != nil {
      continue
    }

    filename := filepath.Base(link)

    if strings.EqualFold(filename, name) {
      ids = append(ids, pid)
    }
  }

  return ids, nil
}

func handleLs() {
  fmt.Println("Files being synced")
  files, err := config.GetFiles()
  if err != nil {
    fmt.Println("Unable to list files")
  }

  if len(files) > 0 {
    fmt.Println()
  }

  for _, f := range files{
    fmt.Printf("%v (%v) => %v\n", f.RemoteName, f.Vendor, f.LocalPath)
  }

  fmt.Println()
}

func handleAdd(args []string) {
  if len(args) < 1 {
    fmt.Println("Arguments given are not compatible")
    fmt.Println("Use \"syncer help add\" to see what arguments to use")
    return
  }
  remote := args[0]
  // TODO: get vendor from flag
  vendor := config.GoogleDrive

  var local string
  if len(args) > 1 {
    local = args[1]
  } else {
    user, err := user.Current()
    if err != nil {
      fmt.Println("Unable to get user info")
      return 
    }

    var vendorDir string 
    switch vendor {
      case config.GoogleDrive:
        vendorDir = "google-drive"
      default:
        vendorDir = ""
    }

    path := filepath.Join(user.HomeDir, "syncer", vendorDir, remote)

    local, err = filepath.Abs(path)
    if err != nil {
      fmt.Println("Unable to get absolute file path for the local file")
      return
    }
  }

  f := config.File{}
  f.RemoteName = remote
  f.LocalPath = local
  // TODO: add flag for vendor choice
  f.Vendor = config.GoogleDrive

  err := config.AddFile(f)
  if err != nil {
    fmt.Println("Unable to add file")
  }
}

func handleRm(args []string) {
  if len(args) < 1 {
    fmt.Println("Arguments given are not compatible")
    fmt.Println("Use \"syncer help rm\" to see what arguments to use")
    return
  }
  name := args[0]

  err := config.RemoveFile(name)
  if err != nil {
    fmt.Println("Unable to remove file")
  }
}

func printGeneralHelp() {
  fmt.Println("Syncer is a cloud file system sync tool to keep remote and local files synced")
  fmt.Println()
  fmt.Println("Usage:")
  fmt.Println()
  fmt.Println("\tsyncer <command> [arguments]")
  fmt.Println()
  fmt.Println("The commands are:")
  fmt.Println()
  fmt.Println("\tauth\tauthenticates to access the remote files")
  fmt.Println("\tstatus\tchecks the status")
  fmt.Println("\tbrowse\tbrowse/list the remote vendor")
  fmt.Println("\tstart\tstarts the syncer daemon")
  fmt.Println("\tstop\tstops the syncer daemon")
  fmt.Println("\tpull\tpulls the latest version of all files or a specific file")
  fmt.Println("\tls\tlists the synced files")
  fmt.Println("\tadd\tadds a file to be synced")
  fmt.Println("\trm\tremoves a file from syncing")
  fmt.Println("\thelp\tprints the help")
  fmt.Println()
  fmt.Println("Use \"syncer help <command>\" for more information about a command")
  fmt.Println()
}

func printCommandHelp(command string) {
  switch strings.ToLower(command) {
    case "pull":
      fmt.Println("Pulls the latest version for every file or a specific file from the vendor")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer pull")
      fmt.Println()
    case "browse":
      fmt.Println("Browses or lists the remote file server")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer browse")
      fmt.Println()
    case "auth":
      fmt.Println("Authenticates for syncing to the cloud file management")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer auth")
      fmt.Println()
    case "status":
      fmt.Println("Prints the current status of the daemon and files synced")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer status")
      fmt.Println()
    case "start":
      fmt.Println("Starts the syncer daemon")
    case "stop":
      fmt.Println("Stops the syncer daemon")
    case "ls":
      fmt.Println("Prints the files being synced")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer ls")
      fmt.Println()
    case "add":
      fmt.Println("Adds the given file to be synced")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer add <remote name> [local path]")
      fmt.Println()
      fmt.Println("Arguments:")
      fmt.Println()
      fmt.Println("\t- <remote name> should be a unique name together with the vendor. ")
      fmt.Println("\t  I.e. every filename/filepath should be unique for the vendor. This is normal for any filesystem")
      fmt.Println("\t- [local path] is an optional argument. The default is $HOME/syncer/$VENDORDIR/RemoteName")
      fmt.Println()
    case "rm":
      fmt.Println("Removes the given file from syncing")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer rm <remote name>")
      fmt.Println()
    case "help":
      fmt.Println("Prints the help")
    default:
      fmt.Println("Unknown command")
  }
}
