package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/icrowley/fake"
	log "github.com/sirupsen/logrus"
)

func main() {
	http.HandleFunc("/zipcode", ZipHandler)
	log.Fatal(http.ListenAndServe(":3113", nil))
}

func ZipHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var z Zipcode
	if err := dec.Decode(&z); err != nil {
		log.Errorf("error in decoder")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	pause := rand.Intn(3000)
	time.Sleep(time.Duration(pause) * time.Millisecond)
	json.NewEncoder(w).Encode(newRandomZiprich(z))
}

type Zipcode struct {
	Zipcode   string `json:"zipcode"`
	Timestamp int64  `json:"timestamp"`
}

type Ziprich struct {
	Zipcode
	Addresses `json:"addresses"`
}

func newRandomZiprich(z Zipcode) Ziprich {
	return Ziprich{
		Zipcode: Zipcode{
			Zipcode:   z.Zipcode,
			Timestamp: time.Now().UnixNano(),
		},
		Addresses: newAddresses(z.Zipcode),
	}
}

type Addresses []Address

func newAddresses(zipcode string) Addresses {
	addresses := make(Addresses, 0)
	for i := 0; i < rand.Intn(10); i++ {
		addresses = append(addresses, newFakeAddress(zipcode))
	}
	return addresses
}

type Address struct {
	Zipcode string `json:"zipcode"`
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

func newFakeAddress(zipcode string) Address {
	return Address{
		Zipcode: zipcode,
		Street:  fake.Street(),
		City:    fake.City(),
		State:   fake.State(),
		Country: "US",
	}
}
