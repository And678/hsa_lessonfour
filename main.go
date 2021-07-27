package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Request struct {
	id int
	createdAt time.Time
	userAgent string
	body string
}

var dbConnection *sql.DB
var cache []byte
var expiry time.Time

func main() {
	var err error
	dbConnection, err = sql.Open("mysql", "root:qwerty@/main")

	if err != nil {
		println(err.Error())
		panic(err)
	}
	defer dbConnection.Close()

	http.HandleFunc("/no-cache", getNotCachedData)
	http.HandleFunc("/cache", getCachedData)
	server := http.Server{Addr: ":4200", ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second}
	server.ListenAndServe()
}

func getData(r *http.Request) []byte {
	dbConnection.Exec("INSERT INTO main.requests (createdAt, userAgent, body) VALUES (?, ?, ?)",
		time.Now(), r.UserAgent(), "r.Body")

	rows, _ := dbConnection.Query("SELECT * FROM main.requests")
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		req := Request{}
		rows.Scan(&req.id, &req.createdAt, &req.userAgent, &req.body)
		requests = append(requests, req)

	}

	return []byte(strconv.Itoa(len(requests)))
}

func getCachedData(w http.ResponseWriter, r *http.Request) {
	if cache == nil || time.Now().Add(-time.Second * time.Duration(math.Log(rand.Float64()))).After(expiry) {
		cache = getData(r)
		expiry = time.Now().Add(time.Second)
	}
	w.Write(cache)
}

func getNotCachedData(w http.ResponseWriter, r *http.Request) {
	resp := getData(r)
	w.Write(resp)
}
