package main

import (
  "log"
  "os"

  "github.com/lpernett/godotenv"
)

func main() {
  godotenv.Load()
  graphDir := os.Getenv("GRAPH_DIR")
  log.Println("GRAPH_DIR:", graphDir)
}
