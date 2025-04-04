package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"syncer/internal"
)

const (
  DAEMON = "syncerd"
)

func main() {
  if runtime.GOOS == "windows" {
    fmt.Println("Windows is not yet implemented")
    os.Exit(1)
  }

  err := internal.Init()
  if err != nil {
    fmt.Printf("Unable to initialize the configuration\n")
    fmt.Printf("%v\n", err)
    os.Exit(1)
  }

  if len(os.Args) == 1 {
    printGeneralHelp()
    os.Exit(0)
  }

  // TODO: handle by flag
  var vendor internal.Vendor = internal.GoogleDrive

  handleCommand(os.Args[1], vendor, os.Args[2:]..., )
}

func handleCommand(command string, vendor internal.Vendor, args ...string) {
  switch strings.ToLower(command) {
    case "pull":
      handlePull(vendor)
    case "push":
      handlePush(vendor)
    case "browse":
      handleBrowse(vendor)
    case "auth":
      handleAuth(vendor)
    case "status":
      handleStatus()
    case "start":
      handleStart()
    case "stop":
      handleStop()
    case "ls":
      handleLs()
    case "add":
      handleAdd(vendor, args)
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

func handleStart() {
  cmd := exec.Command("nohup", DAEMON)
  err := cmd.Start()
  if err != nil {
    log.Fatalf("Unable to start syncer\n%v\n", err)
  }
  fmt.Println("Started the syncer daemon")
}

func handleStop() {
  pids, err := findPid(DAEMON)
  if err != nil {
    fmt.Printf("Unable to find the process for syncer daemon\n%v\n", err)
  }

  for _, pid := range pids {
    cmd := exec.Command("kill", strconv.Itoa(pid))
    err := cmd.Run()
    if err != nil {
      log.Fatalf("Unable to stop syncer\n%v\n", err)
    }
    fmt.Println("Stopped the syncer daemon")
  }
}

func handlePush(vendor internal.Vendor) {
  files, err := internal.GetFiles()
  if err != nil {
    log.Fatalf("Unable to get the watched files\n%v\n", err)
  }

  repo := vendor.Repository()
  err = repo.PushAll(files)
  if err != nil {
    fmt.Printf("Unable to pull files from %v\n%v\n", vendor, err)
  }
}

func handlePull(vendor internal.Vendor) {
  files, err := internal.GetFiles()
  if err != nil {
    log.Fatalf("Unable to get the watched files\n%v\n", err)
  }

  repo := vendor.Repository()
  err = repo.PullAll(files)
  if err != nil {
    fmt.Printf("Unable to pull files from %v\n%v\n", vendor, err)
  }
}

func handleAuth(vendor internal.Vendor) {
  err := vendor.Repository().Authenticate()
  if err != nil {
    log.Fatalf("Error authenticating google drive\n%v\n", err)
  }
}

func handleStatus() {
  ids, err := findPid("syncerd")
  if err != nil {
    fmt.Println("Unable to check if the daemon is running")
  }
  if len(ids) > 0 {
    fmt.Println("Syncer is \033[0;32mrunning\033[0;37m")
  } else {
    fmt.Println("Syncer has \033[0;31mstopped\033[0;37m")
  }

  fmt.Println()
  if internal.GoogleDrive.Repository().IsAuthenticated() {
    fmt.Printf("Google Drive is \033[0;32mauthenticated\033[0;37m\n")
  } else {
    fmt.Printf("Google Drive is \033[0;31mnot authenticated\033[0;37m\n")
  }


  files, err := internal.GetFiles()
  if err != nil {
    fmt.Println("Unable to list files")
  }

  if len(files) > 0 {
    fmt.Println()
  }

  for _, f := range files{
    switch f.Status {
      case internal.Error:
        fmt.Printf("%v (%v): \033[0;31m%v\033[0;37m\n", f.RemoteName, f.Vendor, f.Status)
      case internal.Synced:
        fmt.Printf("%v (%v): \033[0;32m%v\033[0;37m\n", f.RemoteName, f.Vendor, f.Status)
      default:
        fmt.Printf("%v: Unknown\n", f.RemoteName)
    }
  }
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

func handleBrowse(vendor internal.Vendor) {
  files, err := vendor.Repository().List() 
  if err != nil {
    log.Fatalf("Unable to get remote files from %v\n%v", vendor, err)
  }
  fmt.Printf("Files in %v\n\n", vendor)
  for _, filename := range files {
    fmt.Printf("%v\n", filename)
  }
}

func handleLs() {
  fmt.Println("Files being watched")
  files, err := internal.GetFiles()
  if err != nil {
    fmt.Println("Unable to list files")
  }

  if len(files) > 0 {
    fmt.Println()
  }

  for _, f := range files{
    fmt.Printf("%v (%v) => %v\n", f.RemoteName, f.Vendor, f.LocalPath)
  }
}

func handleAdd(vendor internal.Vendor, args []string) {
  if len(args) < 1 {
    fmt.Println("Arguments given are not compatible")
    fmt.Println("Use \"syncer help add\" to see what arguments to use")
    return
  }
  remote := args[0]

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
      case internal.GoogleDrive:
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

  id, err := vendor.Repository().GetRemoteId(remote) 
  if err != nil {
    log.Fatalf("Unable to find remote file")
  }

  f := internal.File{
    RemoteId: id,
    RemoteName: remote,
    LocalPath: local,
    Vendor: internal.GoogleDrive,
  }

  err = internal.AddFile(f)
  if err != nil {
    log.Fatalf("Unable to add file")
  }
}

func handleRm(args []string) {
  if len(args) < 1 {
    fmt.Println("Arguments given are not compatible")
    fmt.Println("Use \"syncer help rm\" to see what arguments to use")
    return
  }
  name := args[0]

  err := internal.RemoveFile(name)
  if err != nil {
    fmt.Println("Unable to remove file")
  }
}

func printGeneralHelp() {
  fmt.Println("Syncer is a cloud file system sync tool to keep remote and local files synchronised")
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
  fmt.Println("\tpush\tpushes the local changes the the remote repository")
  fmt.Println("\tls\tlists the watched files")
  fmt.Println("\tadd\tadds a file to be watched")
  fmt.Println("\trm\tremoves a file from syncing")
  fmt.Println("\thelp\tprints the help")
  fmt.Println()
  fmt.Println("Use \"syncer help <command>\" for more information about a command")
}

func printCommandHelp(command string) {
  switch strings.ToLower(command) {
    case "pull":
      fmt.Println("Pulls the latest version for every file or a specific file from the vendor")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer pull")
    case "push":
      fmt.Println("Pushes the local version to the remote repository for the vendor")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer push")
    case "browse":
      fmt.Println("Browses or lists the remote file server")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer browse [vendor]")
    case "auth":
      fmt.Println("Authenticates for syncing to the cloud file management")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer auth")
    case "status":
      fmt.Println("Prints the current status of the daemon and files being watched")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer status")
    case "start":
      fmt.Println("Starts the syncer daemon")
    case "stop":
      fmt.Println("Stops the syncer daemon")
    case "ls":
      fmt.Println("Prints the files being watched")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer ls")
    case "add":
      fmt.Println("Adds the given file to be watched")
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
    case "rm":
      fmt.Println("Removes the given file from syncing")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer rm <remote name>")
    case "help":
      fmt.Println("Prints the help")
    default:
      fmt.Println("Unknown command")
  }
}
