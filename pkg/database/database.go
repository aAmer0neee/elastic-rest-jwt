package database

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type (
	// https://mholt.github.io/json-to-go/
	EsResponse struct {
		Hits struct {
			Total struct {
				Value    int    `json:"value"`
				Relation string `json:"relation"`
			} `json:"total"`
			MaxScore float64 `json:"max_score"`
			Hits     []struct {
				Index  string  `json:"_index"`
				ID     string  `json:"_id"`
				Score  float64 `json:"_score"`
				Source struct {
					Name     string `json:"name"`
					Address  string `json:"address"`
					Phone    string `json:"phone"`
					Location struct {
						Lat float64 `json:"lat"`
						Lon float64 `json:"lon"`
					} `json:"location"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	Place struct {
		Name     string   `json:"name"`
		Address  string   `json:"address"`
		Phone    string   `json:"phone"`
		Location Location `json:"location"`
	}

	Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	Store interface {
		// returns a list of items, a total number of hits and (or) an error in case of one
		GetPlaces(limit int, offset int) ([]Place, int, error)
		GetNearestPlaces(limit int, lat, lon float64) ([]Place, error)

		CreateNewIndex(string) error
		BukIndexing(string) error
		ChangeIndexSizeSetting(elasticURL,indexName string) error
	}

	EsClient struct {
		client *elasticsearch.Client
	}
)

const (
	dataSetPath = "./data/data.csv"
	schemaPath  = "./configs/schema.json"
)

// создает нового клиента с конфигурацией из CFG
func CreateNewClient(elasticURL string) (*EsClient, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{elasticURL},
	})

	if err != nil {
		return nil, fmt.Errorf("CreateNewClient \"%v\"", err)
	}

	return &EsClient{client: client}, nil
}

// cоздание нового индекса с маппингом из JSON
func (e *EsClient) CreateNewIndex(indexName string) error {

	mapping, err := GetMapShcemaFromJson(schemaPath)
	if err != nil {
		return err
	}
	elasticResponse, err := e.client.Indices.Create(indexName, //	ожидается Create(Имя индекса, Опции запроса (WithBody(io.Reader))
		e.client.Indices.Create.WithBody(mapping),
	)
	if err != nil || elasticResponse.IsError() {

		return fmt.Errorf("CreateNewIndex \"%v\" response from Elasticsearch: \"%v\"", err, elasticResponse.String())
	}
	defer elasticResponse.Body.Close()

	return nil
}

// массовая индексация данных
func (e *EsClient) BukIndexing(indexName string) error {
	jsonData, err := GetDataFromCsv(indexName)
	if err != nil {
		return err
	}

	bulkRes, err := e.client.Bulk( //	Выполняет несколько операций индексации или удаления в одном вызове API
		bytes.NewReader(jsonData),
	)

	if err != nil {
		return fmt.Errorf("BukIndexing \"%v\" status \"%v\"", err, bulkRes.StatusCode)
	}
	return nil
}

// изменение настроек elastic для отображения всех документов индекса dafault: 10 000
func (e *EsClient) ChangeIndexSizeSetting(elasticURL,indexName string) error {

	responseSize := `{
		"index": {
			"max_result_window": 15000
		}
	}`

	elasticSettingsUrl := fmt.Sprintf("%s/%s/_settings",elasticURL, indexName)

	request, err := http.NewRequest(http.MethodPut, elasticSettingsUrl, bytes.NewReader([]byte(responseSize)))
	if err != nil {
		return fmt.Errorf("ChangeIndexSizeSetting \"%v\"", err)
	}

	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return fmt.Errorf("ChangeIndexSizeSetting \"%v\" client response: \"%v\"", err, response)
	}
	defer response.Body.Close()

	return nil
}

// получает io.Reader из JSON
func GetMapShcemaFromJson(path string) (*bytes.Reader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("GetMapShcemaFromJson \"%v\"", err)
	}
	return bytes.NewReader(data), nil
}

// индексация данных из CSV в ElasticSearch
func GetDataFromCsv(indexName string) ([]byte, error) {

	file, err := os.Open(dataSetPath)
	if err != nil {
		return nil, fmt.Errorf("GetDataFromCsv \"%v\"", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t' //	разделитель для ридера

	var buf bytes.Buffer

	for {
		record, err := reader.Read() //	Возвращает слайс string, что позволяет индексировать все поля
		if err != nil {
			if err == io.EOF {

				return buf.Bytes(), nil

			}

			return nil, fmt.Errorf("GetDataFromCsv \"%v\"", err)
		}

		Lon, _ := strconv.ParseFloat(record[4], 64)
		Lat, _ := strconv.ParseFloat(record[5], 64)

		data := Place{
			Name:    record[1],
			Address: record[2],
			Phone:   record[3],
			Location: Location{
				Lat: Lat,
				Lon: Lon,
			},
		}

		jsonRes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("GetDataFromCsv \"%v\"", err)
		}
		// JSON строка с метаданными, которая описывает индекс в Elasticsearch
		buf.WriteString(fmt.Sprintf(`{ "index": { "_index": "%s", "_id": "%s" } }`+"\n", indexName, record[0]))
		buf.Write(jsonRes)
		buf.WriteString("\n")
	}
}

// формирование и отправка запроса к ElasticSearch на получение всех записей
func (e *EsClient) GetPlaces(limit int, offset int) ([]Place, int, error) {

	query := fmt.Sprintf(`{
        "from": %d,
        "size": %d,
		"track_total_hits": true,
        "query": {
            "match_all": {}
        }
    }`, offset, limit)

	res, err := e.client.Search(
		e.client.Search.WithBody(strings.NewReader(query)),
	)

	if err != nil {
		return nil, 0, fmt.Errorf("GetPlaces \"%v\"", err)
	}
	defer res.Body.Close()

	decode, err := DecodeJson(res.Body)
	if err != nil {
		return nil, 0, err
	}

	places := make([]Place, len(decode.Hits.Hits))
	for i, hit := range decode.Hits.Hits {
		places[i] = Place{
			Name:    hit.Source.Name,
			Address: hit.Source.Address,
			Phone:   hit.Source.Phone,
			Location: Location{
				Lat: hit.Source.Location.Lat,
				Lon: hit.Source.Location.Lon,
			},
		}

	}
	return places, decode.Hits.Total.Value, nil
}

// формирование и отправка запроса к ElasticSearch с сортировкой записей по координатам
func (e *EsClient) GetNearestPlaces(limit int, lat, lon float64) ([]Place, error) {
	query := fmt.Sprintf(`{
		"size": %d,
		"track_total_hits": true,
		"sort": [
			{
				"_geo_distance": {
					"location": {
						"lat": %f,
						"lon": %f
					},
					"order": "asc",
					"unit": "km",
					"mode": "min",
					"distance_type": "arc",
					"ignore_unmapped": true
				}
			}
		]
	}`, limit, lat, lon)

	elasticResponse, err := e.client.Search(
		e.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, fmt.Errorf("GetNearestPlaces \"%v\" response from Elasticsearch: \"%v\"", err, elasticResponse.String())
	}
	defer elasticResponse.Body.Close()

	decode, err := DecodeJson(elasticResponse.Body)
	if err != nil {
		return nil, err
	}

	places := make([]Place, len(decode.Hits.Hits))
	for i, hit := range decode.Hits.Hits {
		places[i] = Place{
			Name:    hit.Source.Name,
			Address: hit.Source.Address,
			Phone:   hit.Source.Phone,
			Location: Location{
				Lat: hit.Source.Location.Lat,
				Lon: hit.Source.Location.Lon,
			},
		}

	}
	return places, nil
}

func DecodeJson(r io.ReadCloser) (EsResponse, error) {
	var res EsResponse
	err := json.NewDecoder(r).Decode(&res)
	if err != nil {
		return EsResponse{}, fmt.Errorf("DecodeJson: \"%v\"", err)
	}
	return res, nil
}
