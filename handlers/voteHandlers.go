package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vikks15/ForumAPI/structs"
)

func (env *Env) CreateVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var newVote structs.Vote
	var votedThread structs.Thread
	json.NewDecoder(r.Body).Decode(&newVote) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil
	sqlStatement := ""

	tx, err := env.DB.Begin()
	if err != nil {
		fmt.Print(err)
		fmt.Print("\n")
		return
	}

	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil {
		sqlStatement = `SELECT id FROM thread WHERE id = $1`
	} else {
		sqlStatement = `SELECT id FROM thread WHERE slug = $1`
	}

	row := env.DB.QueryRow(sqlStatement, vars["slug_or_id"])
	scanErr := row.Scan(&newVote.ThreadId)

	if scanErr != nil {
		fmt.Print("\n CreateVote threadNotFound:")
		fmt.Print(scanErr)
		fmt.Print("\n")
		tx.Rollback()

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
	_, err = env.DB.Exec(sqlStatement, newVote.Nickname, newVote.Voice, newVote.ThreadId)

	if err != nil {
		fmt.Print(err)
		tx.Rollback()

		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user or thread\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}

	sqlStatement = `UPDATE thread SET votes = 
					(select sum(voice) from vote where threadId = $1)
					where id = $1`

	_, err = env.DB.Exec(sqlStatement, newVote.ThreadId)

	if err != nil {
		fmt.Print("update votedThreadErr: ")
		fmt.Print(err)
		fmt.Print("\n")
		tx.Rollback()
	} else {
		row := env.DB.QueryRow("SELECT id, title, author, forum, message, votes, slug, created FROM vote JOIN thread ON (threadId = id) where id = " + strconv.Itoa(newVote.ThreadId))
		scanErr := row.Scan(&votedThread.Id, &votedThread.Title, &votedThread.Author, &votedThread.Forum, &votedThread.Message, &votedThread.Votes, &votedThread.Slug, &votedThread.Created)

		if scanErr != nil {
			fmt.Print("votedThread ScanErr: ")
			fmt.Print(scanErr)
			fmt.Print("\n")
		}

		tx.Commit()
		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(votedThread)
		w.Write(response)
	}
}
