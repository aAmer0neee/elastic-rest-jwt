package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aAmer0neee/elastic-rest-jwt/pkg/authorisation"
	"github.com/aAmer0neee/elastic-rest-jwt/pkg/database"
)

var (
	applicationPort = ":" + os.Getenv("APP_PORT")

	elasticURL = "http://" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT")

	indexHtml = "./pkg/configs/index.html"
)

type (
	PageData struct {
		Name   string           `json:"name"`
		Total  int              `json:"total"`
		Places []database.Place `json:"places"`
		Prev   int              `json:"prev"`
		Next   int              `json:"next"`
		Last   int              `json:"last"`
		Page   int              `json:"page"`
	}
)

func Run() {
	
	EsClient, err := database.CreateNewClient(elasticURL)

	if err != nil {
		log.Fatalf("error:\t\"%v\"", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/create/", func(w http.ResponseWriter, r *http.Request) { // PUT http://localhost:8888/create/?name=<index_name>
		CreateIndexHandler(EsClient, w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // GET http://localhost:8888/?page=1
		XmlHandler(EsClient, w, r)
	})

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { // GET http://localhost:8888/api/?page=1
		JsonApiHandler(EsClient, w, r)
	})

	mux.Handle("/api/recommend/", AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // GET http://localhost:8888/api/recommend/?lat=55.674&lon=37.666
		NearestApiHandler(EsClient, w, r) // with header <	Authorization: Bearer <token>	>
	})))

	mux.HandleFunc("/api/get_token/", func(w http.ResponseWriter, r *http.Request) { //	GET http://localhost:8888/api/get_token
		AuthHandler(w, r)
	})

	if err := http.ListenAndServe(applicationPort, mux); err != nil {
		log.Fatalf("error starting server at :8888")
	}
}

// промежуточный хендлер, проверяет JWT токен,
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if sign, err := authorisation.VerifyToken(r); !sign {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r) // если ок, передаем в слудующий хендлер
	})
}

// хендлер для получения токена
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := authorisation.GetToken(w, r)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(token)
}

// хендлер создет индекс -> заполняет с BULK API -> меняет настройи отображения данных
func CreateIndexHandler(store database.Store, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {

		indexName := r.URL.Query().Get("name")
		if indexName == "" {
			http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
			return
		}
		if err := store.CreateNewIndex(indexName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := store.BukIndexing(indexName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := store.ChangeIndexSizeSetting(indexName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "index created", http.StatusCreated)
	} else {
		http.Error(w, "no valid method", http.StatusBadRequest)
	}

}

// хендлер обрабатывает XML с пагинацией
func XmlHandler(store database.Store, w http.ResponseWriter, r *http.Request) {

	limit := 10

	pageParam := r.URL.Query().Get("page")

	var offset int

	if pageParam != "" {
		page, err := strconv.Atoi(pageParam)
		if err != nil {
			http.Error(w, "Invalid param 'page'", http.StatusBadRequest)
			return
		}

		offset = (page - 1) * limit
		data, hits, err := store.GetPlaces(limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		response := PageData{
			Name:   "Places",
			Total:  hits,
			Places: data,
			Prev:   max(0, page-1),
			Next:   page + 1,
			Last:   (hits + limit - 1) / limit,
			Page:   page,
		}

		if response.Page < 1 || response.Page > response.Last {
			http.Error(w, "Invalid value of param 'page'", http.StatusBadRequest)
			return
		}

		tmp, err := template.ParseFiles(indexHtml)

		if err != nil {
			http.Error(w, "Error parse mapping", http.StatusBadRequest)
			return
		}

		tmp.Execute(w, response)
	}

}

// хендлер возвращает JSON API
func JsonApiHandler(client database.Store, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit := 10

	pageParam := r.URL.Query().Get("page")

	var offset int

	if pageParam != "" {
		page, err := strconv.Atoi(pageParam)
		if err != nil {
			http.Error(w, "invalid 'page' param", http.StatusBadRequest)
			return
		}

		offset = (page - 1) * limit
		data, hits, err := client.GetPlaces(limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		response := PageData{
			Name:   "Places",
			Total:  hits,
			Places: data,
			Prev:   max(0, page-1),
			Next:   page + 1,
			Last:   (hits + limit - 1) / limit,
			Page:   page,
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if response.Page < 1 || response.Page > response.Last {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{
				"error": fmt.Sprintf("Invalid '%s' value: '%d'", pageParam, response.Page),
			})
			return
		}
		encoder.Encode(response)

	}

}

func NearestApiHandler(client database.Store, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit := 3

	lonParam := r.URL.Query().Get("lon")
	latParam := r.URL.Query().Get("lat")
	if latParam == "" || lonParam == "" {
		http.Error(w, "The parameters 'lat' and 'lon' are required", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonParam, 64)
	if err != nil {
		http.Error(w, "invalid 'lon' param", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latParam, 64)
	if err != nil {
		http.Error(w, "invalid 'lat' param", http.StatusBadRequest)
		return
	}

	data, err := client.GetNearestPlaces(limit, lat, lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	response := struct {
		Name   string           `json:"name"`
		Places []database.Place `json:"places"`
	}{
		Name:   "Recommendation",
		Places: data,
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)

}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
