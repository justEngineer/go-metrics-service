package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	storage "github.com/justEngineer/go-metrics-service/internal"
	server "github.com/justEngineer/go-metrics-service/internal/http/server"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetric(t *testing.T) {
	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method       string
		url          string
		expectedCode int
	}{
		{method: http.MethodPost, url: "http://localhost:8080/update/gauge/temp/36.6", expectedCode: http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.url, nil)
			w := httptest.NewRecorder()

			urlPart := strings.Split(r.URL.Path, "/")
			idx := slices.IndexFunc(urlPart, func(c string) bool { return c == "update" })
			typeIdx := idx + 1
			nameIdx := idx + 2
			valueIdx := idx + 3
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", urlPart[typeIdx])
			rctx.URLParams.Add("name", urlPart[nameIdx])
			rctx.URLParams.Add("value", urlPart[valueIdx])

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			MetricStorage := storage.New()
			config := server.ServerConfig{Endpoint: "", PathToHTMLTemplate: server.DefaultPathToHTMLTemplate}
			ServerHandler := server.New(MetricStorage, &config)
			// вызовем хендлер как обычную функцию, без запуска самого сервера
			ServerHandler.UpdateMetric(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			//assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetric(t *testing.T) {
	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method          string
		url             string
		expectedCode    int
		expectedCounter int64
		expectedGauge   float64
	}{
		{method: http.MethodGet, url: "http://localhost:8080/value/gauge/temp/",
			expectedCode: http.StatusOK, expectedCounter: int64(0), expectedGauge: float64(36.6)},
		{method: http.MethodGet, url: "http://localhost:8080/value/counter/timeoutInterval/",
			expectedCode: http.StatusOK, expectedCounter: int64(10), expectedGauge: float64(0)},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.url, nil)
			w := httptest.NewRecorder()

			urlPart := strings.Split(r.URL.Path, "/")
			idx := slices.IndexFunc(urlPart, func(c string) bool { return c == "value" })
			typeIdx := idx + 1
			nameIdx := idx + 2
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", urlPart[typeIdx])
			rctx.URLParams.Add("name", urlPart[nameIdx])
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			MetricStorage := storage.New()
			MetricStorage.Gauge["temp"] = 36.6
			MetricStorage.Counter["timeoutInterval"] = 10
			config := server.ServerConfig{Endpoint: "", PathToHTMLTemplate: server.DefaultPathToHTMLTemplate}
			ServerHandler := server.New(MetricStorage, &config)
			// вызовем хендлер как обычную функцию, без запуска самого сервера
			ServerHandler.GetMetric(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			data, err := io.ReadAll(w.Body)
			require.Equal(t, err, nil)
			if slices.IndexFunc(urlPart, func(c string) bool { return c == "gauge" }) >= 0 {
				value, err := strconv.ParseFloat(string(data), 64)
				require.Equal(t, err, nil)
				assert.Equal(t, tc.expectedGauge, value, "Код ответа не совпадает с ожидаемым")
			} else if slices.IndexFunc(urlPart, func(c string) bool { return c == "counter" }) >= 0 {
				value, err := strconv.ParseInt(string(data), 10, 64)
				require.Equal(t, err, nil)
				assert.Equal(t, tc.expectedCounter, value, "Код ответа не совпадает с ожидаемым")
			}
			assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		})
	}
}

func TestMainPage(t *testing.T) {
	MetricStorage := storage.New()
	config := server.ServerConfig{Endpoint: "", PathToHTMLTemplate: server.DefaultPathToHTMLTemplate}

	ServerHandler := server.New(MetricStorage, &config)
	r := httptest.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
	w := httptest.NewRecorder()
	ServerHandler.MainPage(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "Код ответа не совпадает с ожидаемым")
}
