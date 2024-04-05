package server

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"bytes"
	"encoding/json"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	model "github.com/justEngineer/go-metrics-service/internal/models"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

type Handler struct {
	storage *storage.MemStorage
	config  *ServerConfig
}

func New(metricsService *storage.MemStorage, config *ServerConfig) *Handler {
	return &Handler{metricsService, config}
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
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
		val, ok := h.storage.Gauge[name]
		if ok {
			body = strconv.FormatFloat(val, 'f', -1, 64)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if valueType == "counter" {
		val, ok := h.storage.Counter[name]
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
	_, err := w.Write([]byte(body))
	if err != nil {
		panic(err)
	}
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
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
			h.storage.Gauge[name] = value
		} else {
			http.Error(w, "Wrong data type, float64 is expected", http.StatusBadRequest)
			return
		}
	} else if valueType == "counter" {
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err == nil {
			h.storage.Counter[name] += value
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

func (h *Handler) MainPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	urlPart := strings.Split(r.URL.Path, "/")
	if len(urlPart) > 2 {
		http.Error(w, "Wrong URL", http.StatusBadRequest)
		return
	}
	if len(urlPart[1]) != 0 {
		http.Error(w, "Wrong URL", http.StatusBadRequest)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tmplFile := filepath.Join(filepath.Dir(wd), h.config.PathToHTMLTemplate)
	tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
	if err != nil {
		tmplFile := filepath.Join(filepath.Dir(wd), "/go-metrics-service/internal/http/server/main_page_html.tmpl")
		tmpl, err = template.New(tmplFile).ParseFiles(tmplFile)
		w.Header().Set("Content-Encoding", "gzip")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			//			panic(err)
			return
		}
	}
	filename := h.config.PathToHTMLTemplate[strings.LastIndex(h.config.PathToHTMLTemplate, "/")+1:]
	err = tmpl.ExecuteTemplate(w, filename, h.storage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//		panic(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricAsJSON(w http.ResponseWriter, r *http.Request) {
	// if r.Header.Get("Content-Type") != "application/json" {
	// 	logger.Log.Warn("Content-Type should be application/json")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	var requestedMetric model.Metrics
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(r.Body)
	if err != nil {
		logger.Log.Warn("Error parsing request body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(buffer.Bytes(), &requestedMetric); err != nil {
		logger.Log.Warn("Error parsing request body as JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if requestedMetric.MType == "gauge" {
		val, ok := h.storage.Gauge[requestedMetric.ID]
		if ok {
			requestedMetric.Value = &val
		} else {
			logger.Log.Warn("Error parsing request body as JSON", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else if requestedMetric.MType == "counter" {
		val, ok := h.storage.Counter[requestedMetric.ID]
		if ok {
			requestedMetric.Delta = &val
		} else {
			logger.Log.Warn("Error parsing request body as JSON", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(requestedMetric)
	if err != nil {
		logger.Log.Warn("Error converting response body to JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		logger.Log.Warn("Error writing response body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	// if r.Header.Get("Content-Type") != "application/json" {
	// 	logger.Log.Warn("Content-Type should be application/json")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	var requestedMetric model.Metrics
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(r.Body)
	if err != nil {
		logger.Log.Warn("Error parsing request body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(buffer.Bytes(), &requestedMetric); err != nil {
		logger.Log.Warn("Error parsing request body as JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if requestedMetric.MType == "gauge" {
		h.storage.Gauge[requestedMetric.ID] = *requestedMetric.Value
	} else if requestedMetric.MType == "counter" {
		h.storage.Counter[requestedMetric.ID] += *requestedMetric.Delta
		val := h.storage.Counter[requestedMetric.ID]
		requestedMetric.Delta = &val
	} else {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(requestedMetric)
	if err != nil {
		logger.Log.Warn("Error converting response body to JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		logger.Log.Warn("Error writing response body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
