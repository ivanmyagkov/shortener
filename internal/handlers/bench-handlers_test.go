package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
)

type requestStruct struct {
	URL string `json:"url"`
}

func BenchmarkServer_PostURL(b *testing.B) {
	b.Run("PostURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := utils.CreateID(16)
			req, _ := http.NewRequest("POST", "http://localhost:9090/", strings.NewReader("http://www."+id+".com"))
			client := &http.Client{}
			b.StartTimer()
			resp, _ := client.Do(req)
			resp.Body.Close()
		}
	})
}

func BenchmarkServer_PostJSON(b *testing.B) {
	b.Run("PostJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := utils.CreateID(16)
			url := requestStruct{URL: "http://www." + id + ".com"}
			reqBody, _ := json.Marshal(url)
			payload := strings.NewReader(string(reqBody))
			req, _ := http.NewRequest("POST", "http://localhost:9090/api/shorten", payload)
			client := &http.Client{}
			b.StartTimer()
			resp, _ := client.Do(req)
			resp.Body.Close()
		}
	})
}

func BenchmarkServer_PostBatch(b *testing.B) {
	URLs := make([]interfaces.BatchRequest, 0, 100000)

	for i := 0; i < cap(URLs); i++ {
		id := utils.CreateID(16)
		URLs = append(URLs, interfaces.BatchRequest{CorrelationID: "user", OriginalURL: "http://www." + id + ".com"})
	}
	b.ResetTimer()
	b.Run("PostBatch", func(b *testing.B) {
		b.StopTimer()
		reqBody, _ := json.Marshal(URLs)
		payload := strings.NewReader(string(reqBody))
		req, _ := http.NewRequest("POST", "http://localhost:9090/api/shorten/batch", payload)
		client := &http.Client{}
		b.StartTimer()
		resp, _ := client.Do(req)
		resp.Body.Close()
	})
}

func BenchmarkServer_GetURLsByUserID(b *testing.B) {
	b.Run("GetURLsByUserID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "http://localhost:9090/api/user/urls", nil)
			client := &http.Client{}
			b.StartTimer()
			resp, _ := client.Do(req)
			resp.Body.Close()
		}
	})
}

func BenchmarkServer_GetURL(b *testing.B) {
	b.Run("GetURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "http://localhost:9090/c25d48c8b3a03c22", nil)
			client := &http.Client{}
			b.StartTimer()
			resp, _ := client.Do(req)
			resp.Body.Close()
		}
	})
}

func BenchmarkServer_DelURLsBATCH(b *testing.B) {

	URLs := make([]string, 0, 100000)

	for i := 0; i < cap(URLs); i++ {
		URLs = append(URLs, "07175c915e801d21")
	}
	b.ResetTimer()
	b.Run("DelURLsBATCH", func(b *testing.B) {
		b.StopTimer()
		reqBody, _ := json.Marshal(URLs)
		payload := strings.NewReader(string(reqBody))
		req, _ := http.NewRequest("DELETE", "http://localhost:9090/api/user/urls", payload)
		client := &http.Client{}
		b.StartTimer()
		resp, _ := client.Do(req)
		resp.Body.Close()
	})
}
