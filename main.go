package main

import (
  "fmt"
  "errors"
  "io"
  "net/http"
  "os"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", GetRoot)
  mux.HandleFunc("/hello", GetHello)

  err := http.ListenAndServe(":3333", mux)

  if errors.Is(err, http.ErrServerClosed) {
    fmt.Println("server closed")
  } else if err != nil {
    fmt.Printf("error starting server: %s\n", err)
    os.Exit(1)
  }
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
  fmt.Println("got / request")
  w.WriteHeader(http.StatusOK)
  io.WriteString(w, "This is my website!\n")
}

func GetHello(w http.ResponseWriter, r * http.Request) {
  fmt.Println("got /hello request")
  w.WriteHeader(http.StatusOK)
  io.WriteString(w, "Hello, HTTP!\n")
}
