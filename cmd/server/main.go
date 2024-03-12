package main

import (
	"fmt"

	"slices"
	"strconv"
	"strings"

	//"io"
	"net/http"
	//"net/url"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	// продолжаем обработку запроса
	// ...
}

func main2() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(".."))
	mux.HandleFunc(`/update/`, mainPage)
	mux.Handle(`/golang/`, http.StripPrefix(`/golang/`, fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./main.go")
	})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func main3() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

type MemStorage struct {
	// указаны некоторые поля структуры
	gauge   map[string]float64
	counter map[string]int64
}

func mainPage5(res http.ResponseWriter, req *http.Request) {
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

var MetricStorage MemStorage

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// if r.Header.Get("Content-Type") != "text/plain" {
	// 	http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
	// 	return
	// }
	urlPart := strings.Split(r.URL.Path, "/")
	idx := slices.IndexFunc(urlPart, func(c string) bool { return c == "update" })
	//log.Println(strconv.Itoa(len(urlPart) - idx))
	if len(urlPart)-idx < 4 {
		http.Error(w, "URL is too short", http.StatusNotFound)
		return
	}
	nameIdx := idx + 2
	valueIdx := idx + 3
	if urlPart[idx+1] == "gauge" {
		value, err := strconv.ParseFloat(urlPart[valueIdx], 64)
		if err == nil {
			MetricStorage.gauge[urlPart[nameIdx]] = value
		} else {
			http.Error(w, "Wrong data type, float64 is expected", http.StatusBadRequest)
			return
		}
	} else if urlPart[2] == "counter" {
		value, err := strconv.ParseInt(urlPart[valueIdx], 10, 64)
		if err == nil {
			MetricStorage.counter[urlPart[nameIdx]] += value
		} else {
			http.Error(w, "Wrong data type, int64 is expected", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}
	// устанавливаем заголовок Content-Type
	// для передачи клиенту информации, кодированной в JSON
	w.Header().Set("Content-Type", "text/plain")
	//w.Header().Set("Content-Length", strconv.Itoa(len(r.URL.Path)))
	// устанавливаем код 200
	w.WriteHeader(http.StatusOK)
	// пишем тело ответа
	//w.Write([]byte("Hello"))
}

func main() {
	MetricStorage.gauge = make(map[string]float64)
	MetricStorage.counter = make(map[string]int64)

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
