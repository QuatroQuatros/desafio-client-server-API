package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Dolar struct {
	Valor string `json:"dolar"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var dolar Dolar

	err = json.Unmarshal(body, &dolar)
	if err != nil {
		panic(err)
	}

	fmt.Println(dolar.Valor)

	//io.Copy(os.Stdout, res.Body)

	file, err := os.Create("cotacao.txt")

	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", dolar.Valor))
	if err != nil {
		panic(err)
	}
}
