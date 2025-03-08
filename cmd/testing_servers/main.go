package main

import (
    "fmt"
    "net/http"
)

const fileName = "files/largefile.bin"

func main() {

    http.HandleFunc("/", serveFile)

    fmt.Println("Server running on http://localhost:8080")
    fmt.Println("Download file: http://localhost:8080/")
    http.ListenAndServe(":8080", nil)
}

func serveFile(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, fileName)
}
