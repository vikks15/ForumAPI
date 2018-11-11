package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"


	"github.com/vikks15/ForumAPI/structs"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func createForum(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	var newForum Forum
	json.NewDecoder(r.Body).Decode(&newForum) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	sqlStatement := `INSERT INTO forum (slug, title, usernick) VALUES ($1,$2,$3)`
	_, err = db.Exec(sqlStatement, newForum.Slug, newForum.Title, newForum.User)

	if err == nil {
		//User case check
		row := db.QueryRow("SELECT nickname FROM forumUser WHERE nickname = '" + newForum.User + "'")
		scanErr := row.Scan(&newForum.User)
		if scanErr != nil {
			//fmt.Print(scanErr)
		}

		response, _ := json.Marshal(newForum)
		w.WriteHeader(http.StatusCreated) //201
		w.Write(response)
	} else if err != nil && strings.Contains(err.Error(), "pq: duplicate key") {
		w.WriteHeader(http.StatusConflict) //409
		var existingForum Forum
		row := db.QueryRow("SELECT * FROM forum WHERE slug = '" + newForum.Slug + "'")
		scanErr := row.Scan(&existingForum.Slug, &existingForum.Title, &existingForum.User, &existingForum.Posts, &existingForum.Threads)
		if scanErr != nil {
			//fmt.Print(scanErr)
		}
		response, _ := json.Marshal(existingForum)
		w.Write(response)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id " + newForum.User + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}

func getForumDetails(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	row := db.QueryRow("SELECT * FROM forum WHERE slug = '" + vars["slug"] + "'")
	var currentForum Forum
	err = row.Scan(&currentForum.Slug, &currentForum.Title, &currentForum.User, &currentForum.Posts, &currentForum.Threads)

	if err == nil {
		row := db.QueryRow("SELECT nickname FROM forumUser WHERE nickname = '" + currentForum.User + "'")
		scanErr := row.Scan(&currentForum.User)
		if scanErr != nil {
			fmt.Print("/n getForumDetails get User err: ")
			fmt.Print(scanErr)
		}
		response, _ := json.Marshal(currentForum)
		w.Write(response)
	} else {
		fmt.Print("/ng etForumDetails err: ")
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id" + currentForum.User + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}

func getThreads(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	params := r.URL.Query()
	limit := params.Get("limit")
	since := params.Get("since")
	desc := params.Get("desc")

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	vars := mux.Vars(r)
	forumRecord := ""
	row := db.QueryRow("SELECT title FROM forum WHERE slug = '" + vars["slug"] + "'")
	forumScanErr := row.Scan(&forumRecord)

	if forumScanErr != nil {
		//fmt.Print("\n forumScanErr:")
		//fmt.Print(forumScanErr != nil)

		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find forum by slug: " + vars["slug"] + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}

	var threadsArr []Thread
	var currentThread Thread
	var tmpVotes sql.NullInt64

	sqlStatement := "SELECT * FROM thread WHERE forum = '" + vars["slug"] + "'"

	if since != "" && desc == "true" {
		sqlStatement += " AND created <= '" + since + "' "
	} else if since != "" {
		sqlStatement += " AND created >= '" + since + "' "
	}

	if desc == "true" {
		sqlStatement += " order by created desc "
	} else {
		sqlStatement += " order by created "
	}

	if limit != "" {
		sqlStatement += " limit " + limit
	}

	rows, queryErr := db.Query(sqlStatement)
	if queryErr != nil {
		//fmt.Print("\n queryErr: ")
		//fmt.Print(queryErr)
	}

	rowsNum := 0
	for rows.Next() {
		err = rows.Scan(&currentThread.Id, &currentThread.Title, &currentThread.Author, &currentThread.Forum, &currentThread.Message, &tmpVotes, &currentThread.Slug, &currentThread.Created)
		if !tmpVotes.Valid {
			currentThread.Votes = 0
		}
		if err != nil {
			//fmt.Print("\n soo:")
			//fmt.Print(err)
		}
		rowsNum++
		//currentThread.Id = 42
		threadsArr = append(threadsArr, currentThread)
	}

	if threadsArr == nil {
		var emptyThreadArr [0]Thread
		response, _ := json.Marshal(emptyThreadArr)
		w.Write(response)
	} else {
		response, _ := json.Marshal(threadsArr)
		w.Write(response)
	}
}

func getForumUsers(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	params := r.URL.Query()
	limit := params.Get("limit")
	since := params.Get("since")
	desc := params.Get("desc")

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	vars := mux.Vars(r)
	forumRecord := ""
	qquery := "SELECT title FROM forum WHERE slug = '" + vars["slug_or_id"] + "'"
	row := db.QueryRow(qquery)
	forumScanErr := row.Scan(&forumRecord)

	if forumScanErr != nil {
		fmt.Print("\n \n \n Get users forumScanErr: ")
		fmt.Print("\n" + qquery)
		fmt.Print(forumScanErr)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find forum by slug: " + vars["slug_or_id"] + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}

	var usersArr []User
	var currentUser User
	sinceStatement := ""

	if since != "" && desc == "true" {
		sinceStatement = " AND nickname < '" + since + "' "
	} else if since != "" {
		sinceStatement = " AND nickname > '" + since + "' "
	}

	sqlStatement := `(SELECT DISTINCT U.nickname, U.fullname, U.about, U.email
		FROM thread T JOIN forumUser U ON (author = nickname)
		WHERE forum = $1` + sinceStatement +
		`) UNION
		(SELECT DISTINCT U.nickname, U.fullname, U.about, U.email
		FROM post JOIN forumUser U ON (author = nickname)
		WHERE forum = $2 ` + sinceStatement + `)`

	if desc == "true" {
		sqlStatement += " ORDER BY nickname DESC "
	} else {
		sqlStatement += " ORDER BY nickname "
	}

	if limit != "" {
		sqlStatement += " limit " + limit
	}

	rows, queryErr := db.Query(sqlStatement, vars["slug_or_id"], vars["slug_or_id"])
	if queryErr != nil {
		fmt.Print("\n Get users queryErr: ")
		fmt.Print(queryErr)
	}

	rowsNum := 0
	for rows.Next() {
		err = rows.Scan(&currentUser.Nickname, &currentUser.FullName, &currentUser.About, &currentUser.Email)
		if err != nil {
			fmt.Print("\n Get users rowScan Err:")
			fmt.Print(err)
		}
		rowsNum++
		usersArr = append(usersArr, currentUser)
	}

	if usersArr == nil {
		var emptyUsersArr [0]User
		response, _ := json.Marshal(emptyUsersArr)
		w.Write(response)
	} else {
		response, _ := json.Marshal(usersArr)
		w.Write(response)
	}
}
