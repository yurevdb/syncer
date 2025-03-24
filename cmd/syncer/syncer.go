package main

import (
	"fmt"
	"os"
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
  }

  if len(os.Args) == 1 {
    printGeneralHelp()
    os.Exit(0)
  }

  handleCommand(os.Args[1], os.Args[2:]...)
}

func handleCommand(command string, args ...string) {
  switch strings.ToLower(command) {
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
  fmt.Println("Authenticating...")
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

  fmt.Println()
  files, err := config.GetFiles()
  if err != nil {
    fmt.Println("Unable to list files")
  }
  for _, f := range files{
    switch f.Status {
      case config.Error:
        fmt.Printf("%v: \033[0;31m%v\033[0;37m\n", f.RemoteName, f.Status)
      case config.Synced:
        fmt.Printf("%v: \033[0;32m%v\033[0;37m\n", f.RemoteName, f.Status)
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
  fmt.Println()
  files, err := config.GetFiles()
  if err != nil {
    fmt.Println("Unable to list files")
  }
  for _, f := range files{
    fmt.Printf("%v\n", f.RemoteName)
  }
  fmt.Println()
}

func handleAdd(args []string) {
  if len(args) < 2 {
    fmt.Println("Arguments given are not compatible")
    fmt.Println("Use \"syncer help add\" to see what arguments to use")
    return
  }
  remote := args[0]
  local := args[1]

  f := config.File{}
  f.RemoteName = remote
  f.LocalPath = local

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
  fmt.Println("\tstart\tstarts the syncer daemon")
  fmt.Println("\tstop\tstops the syncer daemon")
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
    case "auth":
      fmt.Println("Authenticates for syncing to the cloud file management")
    case "status":
      fmt.Println("Prints the current status of the daemon and files synced")
    case "start":
      fmt.Println("Starts the syncer daemon")
    case "stop":
      fmt.Println("Stops the syncer daemon")
    case "ls":
      fmt.Println("Prints the files being synced")
    case "add":
      fmt.Println("Adds the given file to be synced")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer add [remote name] [local path]")
      fmt.Println()
    case "rm":
      fmt.Println("Removes the given file from syncing")
      fmt.Println()
      fmt.Println("Usage:")
      fmt.Println()
      fmt.Println("\tsyncer rm [remote name]")
      fmt.Println()
    case "help":
      fmt.Println("Prints the help")
    default:
      fmt.Println("Unknown command")
  }
}
