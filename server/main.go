package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type USDBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Cotacao struct {
	USDBRL `json:"USDBRL"`
	gorm.Model
}

// type Cotacao struct {
// 	Code       string `json:"code"`
// 	Codein     string `json:"codein"`
// 	Name       string `json:"name"`
// 	High       string `json:"high"`
// 	Low        string `json:"low"`
// 	VarBid     string `json:"varBid"`
// 	PctChange  string `json:"pctChange"`
// 	Bid        string `json:"bid"`
// 	Ask        string `json:"ask"`
// 	Timestamp  string `json:"timestamp"`
// 	CreateDate string `json:"create_date"`
// 	gorm.Model
// }

var db *gorm.DB
var c *Cotacao

func main() {
	//criando o banco
	conn, err := gorm.Open(sqlite.Open("desafio.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db = conn

	db.AutoMigrate(&Cotacao{})

	//configurando o servidor
	http.HandleFunc("/cotacao", handler)
	fmt.Println("Servidor rodando no endereço: http://127.0.0.1:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//ctx := r.Context()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	log.Println("Request iniciada")
	defer log.Println("Request finalizada")

	select {
	case <-time.After(time.Millisecond * 200):
		cotacao, error := getCotacao()
		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		c = cotacao

		data := map[string]interface{}{
			"dolar": c.Bid,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)

	case <-ctx.Done():
		http.Error(w, "Request cancelada pelo cliente", http.StatusRequestTimeout)
	}

	err := insert_cotacao(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("err")
		return
	}

}

func getCotacao() (*Cotacao, error) {
	resp, error := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if error != nil {
		return nil, error
	}

	defer resp.Body.Close()
	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}

	var c Cotacao

	error = json.Unmarshal(body, &c)
	if error != nil {
		return nil, error
	}

	return &c, nil
}

func insert_cotacao(cotacao *Cotacao) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	tx := db.WithContext(ctx)

	select {
	case <-time.After(time.Millisecond * 10):

		tx.Create(&cotacao)

	case <-ctx.Done():

		return errors.New("Request timeout")
	}
	return nil
}
