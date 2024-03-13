package main

import (
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	storage "github.com/justEngineer/go-metrics-service/internal"
)

func UpdateMetric(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
				storage.Gauge[urlPart[nameIdx]] = value

			} else {
				http.Error(w, "Wrong data type, float64 is expected", http.StatusBadRequest)
				return
			}
		} else if urlPart[2] == "counter" {
			value, err := strconv.ParseInt(urlPart[valueIdx], 10, 64)
			if err == nil {
				storage.Counter[urlPart[nameIdx]] += value
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
}

func main() {
	var MetricStorage storage.MemStorage
	MetricStorage.Gauge = make(map[string]float64)
	MetricStorage.Counter = make(map[string]int64)

	// mux := http.NewServeMux()
	// mux.HandleFunc(`/update/`, mainPage)

	http.HandleFunc("/update/", UpdateMetric(&MetricStorage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
