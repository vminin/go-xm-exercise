package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/vminin/go/xm-exercise/company"
)

const (
	defaultURL = "postgresql://xmusr:xmpass@localhost/xm_db"
	user       = "xmusr"
	pass       = "xmpass"
)

var (
	pgURL   string
	options string
)

func init() {
	flag.StringVar(&pgURL, "pgurl", defaultURL, "PostgreSQL configuration URL")
	flag.StringVar(&options, "opts", "2", "Choose option from exercise: 1 or 2 or both: 1,2 (separated by comma). Default 2")
}

func main() {
	flag.Parse()

	pool, err := pgxpool.Connect(context.Background(), pgURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	options := strings.Split(options, ",")

	model := company.NewModel(pool)

	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/companies", company.List(model))

	create := company.Create(model)
	router.POST("/companies", wrapOptions(create, options))

	router.GET("/companies/:id", company.Find(model))

	delete := company.Delete(model)
	router.DELETE("/companies/:id", wrapOptions(delete, options))

	router.PUT("/companies/:id", company.Update(model))

	log.Fatal(http.ListenAndServe(":8080", router))
}

func wrapOptions(h httprouter.Handle, options []string) httprouter.Handle {
	for _, opt := range options {
		switch opt {
		case "1":
			h = company.CyprusRequest(h)
		case "2":
			h = company.BasicAuth(h, user, pass)
		}
	}
	return h
}
