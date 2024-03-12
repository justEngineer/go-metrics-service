type Request struct {
    // указаны некоторые поля структуры
    Method        string
    URL           *url.URL
    Header        Header
    Body          io.ReadCloser
    ContentLength int64
    Host          string
    // ...
} 

const (
    MethodGet     = "GET"
    MethodHead    = "HEAD"
    MethodPost    = "POST"
    MethodPut     = "PUT"
    MethodPatch   = "PATCH"
    MethodDelete  = "DELETE"
    MethodConnect = "CONNECT"
    MethodOptions = "OPTIONS"
    MethodTrace   = "TRACE"
) 