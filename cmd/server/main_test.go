package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	storage "github.com/justEngineer/go-metrics-service/internal"

	"github.com/stretchr/testify/assert"
)

// func TestUpdateMetric(t *testing.T) {
// 	type want struct {
// 		contentType string
// 		statusCode  int
// 		user        User
// 	}
// 	tests := []struct {
// 		name    string
// 		request string
// 		users   map[string]User
// 		want    want
// 	}{
// 		{
// 			name: "simple test #1",
// 			users: map[string]User{
// 				"id1": {
// 					ID:        "id1",
// 					FirstName: "Misha",
// 					LastName:  "Popov",
// 				},
// 			},
// 			want: want{
// 				contentType: "application/json",
// 				statusCode:  200,
// 				user: User{ID: "id1",
// 					FirstName: "Misha",
// 					LastName:  "Popov",
// 				},
// 			},
// 			request: "/users?user_id=id1",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
// 			w := httptest.NewRecorder()
// 			h := http.HandlerFunc(UpdateMetric(tt.users))
// 			h(w, request)

// 			result := w.Result()

// 			assert.Equal(t, tt.want.statusCode, result.StatusCode)
// 			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

// 			userResult, err := io.ReadAll(result.Body)
// 			require.NoError(t, err)
// 			err = result.Body.Close()
// 			require.NoError(t, err)

// 			err = json.Unmarshal(userResult, &user)
// 			require.NoError(t, err)

// 			assert.Equal(t, tt.want.user, user)
// 		})
// 	}
// }

// package main

// import (
//     "net/http"
//     "net/http/httptest"
//     "testing"

//     "github.com/stretchr/testify/assert"
// )

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

			var MetricStorage storage.MemStorage
			MetricStorage.Gauge = make(map[string]float64)
			MetricStorage.Counter = make(map[string]int64)
			handler := UpdateMetric(&MetricStorage)
			// вызовем хендлер как обычную функцию, без запуска самого сервера
			handler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			//assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		})
	}
}
