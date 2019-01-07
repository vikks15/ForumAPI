package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vikks15/ForumAPI/structs"
)

func (env *Env) CreateVote(w http.ResponseWriter, r *http.Request) {
	var err error
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
	defer tx.Rollback()

	_, err = tx.Exec("SET LOCAL synchronous_commit = OFF")

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

	row := tx.QueryRow(sqlStatement, vars["slug_or_id"])
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

	//---------------------------------------------
	voteInc := 42

	stmt, err := tx.Prepare(`INSERT INTO vote (nickname, voice, threadId)
							VALUES ($1,$2,$3);`)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(newVote.Nickname, newVote.Voice, newVote.ThreadId)

	if (err != nil) && (strings.Contains(err.Error(), "foreign key constraint")) {
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}

	if (err != nil) && (strings.Contains(err.Error(), "unique constraint")) {
		tx.Rollback()
		tx, err = env.DB.Begin()

		if err != nil {
			fmt.Print(err)
			fmt.Print("\n")
			return
		}

		curVoice := 0
		sqlStatement = `SELECT voice FROM vote
						WHERE nickname = $1 AND threadId = $2`

		row := tx.QueryRow(sqlStatement, newVote.Nickname, newVote.ThreadId)
		scanErr = row.Scan(&curVoice)

		if scanErr != nil {
			fmt.Print("CurVoice scanErr: " + scanErr.Error() + "\n")
			return
		}

		if curVoice != newVote.Voice { //change of voice
			sqlStatement = `UPDATE vote SET voice = $1
							WHERE nickname = $2 AND threadId = $3`

			_, err = tx.Exec(sqlStatement, newVote.Voice, newVote.Nickname, newVote.ThreadId)

			if err != nil {
				fmt.Print("update voice err: ")
				fmt.Print(err)
				fmt.Print("\n")
				return
			}

			if newVote.Voice == 1 {
				voteInc = 2
			} else if newVote.Voice == -1 {
				voteInc = -2
			}

		} else { //same voice
			voteInc = 0
		}
	} else if err != nil {
		fmt.Print(err)
		fmt.Print("\n")
		return
	}

	if voteInc == 42 { // new voice inserted
		voteInc = newVote.Voice
	}

	sqlStatement = `UPDATE thread SET votes = votes + $1 WHERE id = $2 RETURNING *`
	row = tx.QueryRow(sqlStatement, voteInc, newVote.ThreadId)
	scanErr = row.Scan(&votedThread.Id, &votedThread.Title, &votedThread.Author, &votedThread.Forum,
		&votedThread.Message, &votedThread.Votes, &votedThread.Slug, &votedThread.Created)

	if scanErr != nil {
		fmt.Print("update Thread ScanErr: ")
		fmt.Print(scanErr)
		fmt.Print("\n")
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
	response, _ := json.Marshal(votedThread)
	w.Write(response)
}
