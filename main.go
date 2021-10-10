package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Coaster struct {
	Name         string `json:"name,"`
	Manufacturer string `json:"manufacturer"`
	ID           string `json:"id,"`
	InPark       string `json:"in_park,"`
	Height 		 int `json:"height,"`
}

type coasterHandlers struct {
	sync.Mutex
	store map[string]Coaster
}

func (h *coasterHandlers) coasters(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
		case "GET":
			h.get(w, r)
			return
		case "POST":
			h.post(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method Not Allowed"))
			return
	}

}

func (h *coasterHandlers) getCoaster(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	coaster, ok := h.store[parts[2]]
	h.Unlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(coaster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coasterHandlers) get(w http.ResponseWriter, r *http.Request) {
	coasters := make([]Coaster, len(h.store))

	h.Lock()
	i := 0
	for _, co := range h.store {
		coasters[i] = co
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(coasters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coasterHandlers) post(w http.ResponseWriter, r *http.Request) {
	
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return;
	}


	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need Content-Type application/json, but got '%s", ct)))
	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return;
	}

	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())

	h.Lock()
	h.store[coaster.ID] = coaster
	h.Unlock()
}

func newCoasterHandlers() *coasterHandlers {
	return &coasterHandlers{
		store: map[string]Coaster{
			"id1":  {
				Name: "Fury 325",
				Height: 99,
				ID: "id1",
				InPark: "Carwowinds",
				Manufacturer: "B+M",
			},
		},
	}
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal{
	password := "pass"
	return &adminPortal{password: password}
}

func (a *adminPortal) handler(w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized"))
		return
	}

	w.Header().Set("Content-Type", "html")
	w.Write([]byte("<html><h1>Super Secret Admin Portal</h1></html>"))

}

func (h *coasterHandlers) getRandomCoaster(w http.ResponseWriter, r *http.Request) {
	ids := make([]string, len(h.store))
	h.Lock()
	i := 0
	for id := range h.store {
		ids[i] = id
		i++
	}
	defer h.Unlock()

	var target string
	if len(ids) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 {
		target = ids[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))]
	}

	w.Header().Add("location", fmt.Sprintf("/coasters/%s", target))
	w.WriteHeader(http.StatusFound)
}

func main() {
	admin := newAdminPortal()
	coasterHandlers := newCoasterHandlers()
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.getCoaster)
	http.HandleFunc("/admin", admin.handler)
	http.HandleFunc("/coasters/random", coasterHandlers.getRandomCoaster)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
	log.Fatal(http.ListenAndServe(":8080", nil))
}
