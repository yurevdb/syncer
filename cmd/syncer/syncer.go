package main

import (
	"fmt"
	"os"
	"runtime"
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
    case "status":
    case "start":
    case "stop":
    case "ls":
    case "add":
    case "rm":
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
    case "rm":
      fmt.Println("Removes the given file from syncing")
    case "help":
      fmt.Println("Prints the help")
    default:
      fmt.Println("Unknown command")
  }
}
