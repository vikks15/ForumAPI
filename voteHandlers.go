package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func createVote(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	vars := mux.Vars(r)
	var newVote Vote
	var votedThread Thread
	json.NewDecoder(r.Body).Decode(&newVote) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil
	sqlStatement := ""

	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil {
		sqlStatement = `SELECT id FROM thread WHERE id = $1`
	} else {
		sqlStatement = `SELECT id FROM thread WHERE slug = $1`
	}

	row := db.QueryRow(sqlStatement, vars["slug_or_id"])
	scanErr := row.Scan(&newVote.ThreadId)

	if scanErr != nil {
		fmt.Print("\n CreateVote threadNotFound:")
		fmt.Print(scanErr)
		fmt.Print("\n")

		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find thread by slug_or_id " + vars["slug_or_id"] + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
		// row1 := db.QueryRow("SELECT id FROM thread ORDER BY id DESC LIMIT 1")
		// scanErr = row1.Scan(&newVote.ThreadId)
	}

	sqlStatement = `INSERT INTO vote (nickname, voice, threadId) VALUES ($1,$2,$3)`
	_, err = db.Exec(sqlStatement, newVote.Nickname, newVote.Voice, newVote.ThreadId)

	if err == nil {
		row := db.QueryRow("SELECT id, title, author, forum, message, votes, slug, created FROM vote JOIN thread ON (threadId = id) where id = " + strconv.Itoa(newVote.ThreadId))
		scanErr := row.Scan(&votedThread.Id, &votedThread.Title, &votedThread.Author, &votedThread.Forum, &votedThread.Message, &votedThread.Votes, &votedThread.Slug, &votedThread.Created)

		if scanErr != nil {
			fmt.Print("votedThreadErr: ")
			fmt.Print(scanErr)
			fmt.Print("\n")
		}

		if newVote.Voice == -1 {
			sqlStatement = `UPDATE thread SET votes = votes-1 where slug = $1`
			//votedThread.Votes--
			votedThread.Votes = 1 //wrong here
		} else if newVote.Voice == 1 {
			sqlStatement = `UPDATE thread SET votes = votes+1 where slug = $1`
			votedThread.Votes++
		}
		row = db.QueryRow(sqlStatement, votedThread.Slug)

		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(votedThread)
		w.Write(response)
	} else {
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id " + newVote.Nickname + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}
