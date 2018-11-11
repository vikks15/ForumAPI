package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vikks15/ForumAPI/structs"

	_ "github.com/lib/pq"
)

func ServiceGetStatus(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		structs.DB_HOST, structs.DB_PORT, structs.DB_USER, structs.DB_PASSWORD, structs.DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	var userCount, forumCount, postCount, threadCount, votesCount int64

	sqlStatement := "SELECT COUNT(*) FROM forumUser"
	row := db.QueryRow(sqlStatement)
	_ = row.Scan(&userCount)

	sqlStatement = "SELECT COUNT(*) FROM forum"
	row = db.QueryRow(sqlStatement)
	_ = row.Scan(&forumCount)

	sqlStatement = "SELECT COUNT(*) FROM post"
	row = db.QueryRow(sqlStatement)
	_ = row.Scan(&postCount)

	sqlStatement = "SELECT COUNT(*) FROM thread"
	row = db.QueryRow(sqlStatement)
	_ = row.Scan(&threadCount)

	sqlStatement = "SELECT COUNT(*) FROM vote"
	row = db.QueryRow(sqlStatement)
	_ = row.Scan(&votesCount)

	w.WriteHeader(http.StatusOK)
	responseStruct := map[string]int64{"user": userCount, "forum": forumCount, "post": postCount, "thread": threadCount, "votes": votesCount}
	response, _ := json.Marshal(responseStruct)
	w.Write(response)
}

func ServiceClear(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		structs.DB_HOST, structs.DB_PORT, structs.DB_USER, structs.DB_PASSWORD, structs.DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	sqlQuery := `
	TRUNCATE TABLE post CASCADE;
	TRUNCATE TABLE forumUser CASCADE;
	TRUNCATE TABLE forum CASCADE;
	TRUNCATE TABLE thread CASCADE;
	TRUNCATE TABLE vote CASCADE;`

	_, err = db.Exec(sqlQuery)
	if err != nil {
		fmt.Print("\n serviceClear err ")
		fmt.Print(err)
	}

	w.WriteHeader(http.StatusOK)
}
