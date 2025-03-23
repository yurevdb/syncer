package main

import (
  "fmt"
  
  "syncer/internal/config"
)

func main() {
  err := config.Init()
  if err != nil {
    fmt.Printf("Unable to initialize the configuration\n")
    fmt.Printf("%v\n", err)
  }
}
