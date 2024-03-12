package main

import (
    "fmt"
    "io"
    "net/http"
)

func WriteHandle(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "1")
    fmt.Fprint(w, "2")
    w.Write([]byte("3"))
}

func mainPage(res http.ResponseWriter, req *http.Request) {
    body := fmt.Sprintf("Method: %s\r\n", req.Method)
    body += "Header ===============\r\n"
    for k, v := range req.Header {
        body += fmt.Sprintf("%s: %v\r\n", k, v)
    }
    body += "Query parameters ===============\r\n"
    for k, v := range req.URL.Query() {
        body += fmt.Sprintf("%s: %v\r\n", k, v)
    }
    res.Write([]byte(body))
}

func main() {
    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir(".."))
    mux.Handle(`/golang/`, http.StripPrefix(`/golang/`, fs))
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./main.go")
    })
    err := http.ListenAndServe(":8080", mux)
    if err != nil {
        panic(err)
    }
}