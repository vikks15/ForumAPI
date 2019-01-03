package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vikks15/ForumAPI/structs"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func CreateVote(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		structs.DB_HOST, structs.DB_PORT, structs.DB_USER, structs.DB_PASSWORD, structs.DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	vars := mux.Vars(r)
	var newVote structs.Vote
	var votedThread structs.Thread
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
	}

	sqlStatement = `INSERT INTO vote (nickname, voice, threadId)
					VALUES ($1,$2,$3)
					ON CONFLICT (nickname, threadId) DO UPDATE
					SET voice = $2`
	_, err = db.Exec(sqlStatement, newVote.Nickname, newVote.Voice, newVote.ThreadId)

	if err != nil {
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user or thread\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}

	sqlStatement = `UPDATE thread SET votes = 
					(select sum(voice) from vote where threadId = $1)
					where id = $1`

	_, err = db.Exec(sqlStatement, newVote.ThreadId)

	if err != nil {
		fmt.Print("update votedThreadErr: ")
		fmt.Print(err)
		fmt.Print("\n")
	} else {
		row := db.QueryRow("SELECT id, title, author, forum, message, votes, slug, created FROM vote JOIN thread ON (threadId = id) where id = " + strconv.Itoa(newVote.ThreadId))
		scanErr := row.Scan(&votedThread.Id, &votedThread.Title, &votedThread.Author, &votedThread.Forum, &votedThread.Message, &votedThread.Votes, &votedThread.Slug, &votedThread.Created)

		if scanErr != nil {
			fmt.Print("votedThread ScanErr: ")
			fmt.Print(scanErr)
			fmt.Print("\n")
		}

		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(votedThread)
		w.Write(response)
	}
}
