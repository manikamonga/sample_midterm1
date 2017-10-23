package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func prometheusHandler() http.Handler {
	return prometheus.Handler()
}

type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people []Person

func GetPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

func GetPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(people)
}

func CreatePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	person.ID = params["id"]
	people = append(people, person)
	json.NewEncoder(w).Encode(people)
}

func DeletePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(people)
}

func funcHandler(histogram *prometheus.HistogramVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		code := 500

		defer func() { // Make sure we record a status.
			duration := time.Since(start)
			histogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(duration.Seconds())
		}()

		code = http.StatusBadRequest
		if r.Method == "GET" {
			code = http.StatusOK
			w.WriteHeader(code)
			body, _ := ioutil.ReadAll(r.Body)

			fmt.Printf("GET \"%s\"\n", string(body))

		} else {
			w.WriteHeader(code)
		}
	}
}

func main() {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "GET_Request",
		Help: "Time taken for GET reqest",
	}, []string{"code"})
	router := mux.NewRouter()

	router.Handle("/metrics", prometheusHandler())
	router.Handle("/hash", funcHandler(histogram))

	prometheus.Register(histogram)

	people = append(people, Person{ID: "1", Firstname: "Nic", Lastname: "Raboy", Address: &Address{City: "Dublin", State: "CA"}})
	people = append(people, Person{ID: "2", Firstname: "Maria", Lastname: "Raboy"})
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/people/{id}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/people/{id}", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people/{id}", DeletePersonEndpoint).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":1234", router))
}
