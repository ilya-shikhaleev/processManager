package main

import (
    "fmt"
    "math/rand"             
    "time"
    "os"
)

func main() {
    rand.Seed(time.Now().UTC().UnixNano())
    duration := rand.Intn(100)
    fmt.Println("start sleep", duration)
    time.Sleep(time.Duration(duration) * time.Millisecond)
    fmt.Println("end sleep")
    os.Exit(2)
}
