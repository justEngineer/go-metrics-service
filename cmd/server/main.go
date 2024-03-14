package main

import (
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	storage "github.com/justEngineer/go-metrics-service/internal"
)

func GetMetric(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// разрешаем только POST-запросы
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		urlPart := strings.Split(r.URL.Path, "/")
		idx := slices.IndexFunc(urlPart, func(c string) bool { return c == "value" })
		if len(urlPart)-idx < 3 {
			http.Error(w, "URL is too short", http.StatusNotFound)
			return
		}
		valueType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		var body string
		if valueType == "gauge" {
			val, ok := storage.Gauge[name]
			if ok {
				body = strconv.FormatFloat(val, 'f', -1, 64)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else if valueType == "counter" {
			val, ok := storage.Counter[name]
			if ok {
				body = strconv.FormatInt(val, 10)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			http.Error(w, "Unknown metric type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		// пишем тело ответа
		w.Write([]byte(body))
	}
}

func UpdateMetric(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			// разрешаем только POST-запросы
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		urlPart := strings.Split(r.URL.Path, "/")
		idx := slices.IndexFunc(urlPart, func(c string) bool { return c == "update" })
		if len(urlPart)-idx < 4 {
			http.Error(w, "URL is too short", http.StatusNotFound)
			return
		}
		valueType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		valueStr := chi.URLParam(r, "value")
		if valueType == "gauge" {
			value, err := strconv.ParseFloat(valueStr, 64)
			if err == nil {
				storage.Gauge[name] = value
			} else {
				http.Error(w, "Wrong data type, float64 is expected", http.StatusBadRequest)
				return
			}
		} else if valueType == "counter" {
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err == nil {
				storage.Counter[name] += value
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
		// устанавливаем код 200
		w.WriteHeader(http.StatusOK)

	}
}

const metricList = `<table>
    <thead>
	<tr></tr>
    </thead>
    <tbody>
	{{with .Gauge}}
	    {{range $name, $value := . }}
            <tr>
				<td>{{ $name }}</td>
				<td>{{ $value }}</td>
            </tr>
        {{end}}
	{{end}}
	{{with .Counter}}
	    {{range $name, $value := . }}
            <tr>
				<td>{{ $name }}</td>
				<td>{{ $value }}</td>
            </tr>
        {{end}}
	{{end}}
    </tbody>
</table>`

func mainPage(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			urlPart := strings.Split(r.URL.Path, "/")
			if len(urlPart) > 2 {
				http.Error(w, "Wrong URL", http.StatusBadRequest)
				return
			}
			if len(urlPart[1]) != 0 {
				http.Error(w, "Wrong URL", http.StatusBadRequest)
				return
			}
			t := template.New("Metrics-template")
			t, err := t.Parse(metricList)
			if err != nil {
				panic(err)
			}
			err = t.Execute(w, storage)
			if err != nil {
				panic(err)
			}
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func main() {
	config := GetServerConfig()

	var MetricStorage storage.MemStorage
	MetricStorage.Gauge = make(map[string]float64)
	MetricStorage.Counter = make(map[string]int64)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Post("/update/{type}/{name}/{value}", UpdateMetric(&MetricStorage))
	r.Get("/value/{type}/{name}", GetMetric(&MetricStorage))
	r.Get("/", mainPage(&MetricStorage))

	port := strings.Split(config.endpoint, ":")
	http.ListenAndServe(":"+port[1], r)
}
