package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (env *Env) ServiceGetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	var userCount, forumCount, postCount, threadCount, votesCount int64

	sqlStatement := "SELECT COUNT(*) FROM forumUser"
	row := env.DB.QueryRow(sqlStatement)
	_ = row.Scan(&userCount)

	sqlStatement = "SELECT COUNT(*) FROM forum"
	row = env.DB.QueryRow(sqlStatement)
	_ = row.Scan(&forumCount)

	sqlStatement = "SELECT COUNT(*) FROM post"
	row = env.DB.QueryRow(sqlStatement)
	_ = row.Scan(&postCount)

	sqlStatement = "SELECT COUNT(*) FROM thread"
	row = env.DB.QueryRow(sqlStatement)
	_ = row.Scan(&threadCount)

	sqlStatement = "SELECT COUNT(*) FROM vote"
	row = env.DB.QueryRow(sqlStatement)
	_ = row.Scan(&votesCount)

	w.WriteHeader(http.StatusOK)
	responseStruct := map[string]int64{"user": userCount, "forum": forumCount, "post": postCount, "thread": threadCount, "votes": votesCount}
	response, _ := json.Marshal(responseStruct)
	w.Write(response)
}

func (env *Env) ServiceClear(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	sqlQuery := `
	TRUNCATE TABLE post CASCADE;
	TRUNCATE TABLE forumUser CASCADE;
	TRUNCATE TABLE forum CASCADE;
	TRUNCATE TABLE thread CASCADE;
	TRUNCATE TABLE vote CASCADE;`

	_, err := env.DB.Exec(sqlQuery)
	if err != nil {
		fmt.Print("\n serviceClear err ")
		fmt.Print(err)
	}

	w.WriteHeader(http.StatusOK)
}
