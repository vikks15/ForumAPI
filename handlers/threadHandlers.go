package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vikks15/ForumAPI/structs"
)

func (env *Env) CreateThread(w http.ResponseWriter, r *http.Request) {
	var err error
	var newThread structs.Thread
	forumUpdateQuery := ""
	json.NewDecoder(r.Body).Decode(&newThread) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	tx, err := env.DB.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("SET LOCAL synchronous_commit = OFF")

	if err != nil {
		fmt.Println(err)
		return
	}

	//---------------User case check-----------------
	row := tx.QueryRow("SELECT nickname FROM forumUser WHERE nickname = '" + newThread.Author + "'")
	scanErr := row.Scan(&newThread.Author)

	if scanErr != nil || newThread.Author == "" {
		fmt.Print("\nUser check in createThread err: ")
		fmt.Print(scanErr)
		w.WriteHeader(http.StatusNotFound) //404
		errorMsg := map[string]string{"message": "Can't find user with id " + newThread.Author + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}
	//------------------------------------------------

	sqlStatement := `INSERT INTO thread (title, author, forum, message, slug, created) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
	row = tx.QueryRow(sqlStatement, newThread.Title, newThread.Author, newThread.Forum, newThread.Message, newThread.Slug, newThread.Created.UTC())
	err = row.Scan(&newThread.Id)

	if err == nil {
		//Forum case check
		row = tx.QueryRow("SELECT slug FROM forum WHERE slug = '" + newThread.Forum + "'")
		scanErr = row.Scan(&newThread.Forum)

		if scanErr != nil {
			fmt.Print("\nForum check in createThread err: ")
			fmt.Println(scanErr)
		}

		forumUpdateQuery = "UPDATE forum SET threads = threads + 1 WHERE slug = '" + newThread.Forum + "'"
		_, err = tx.Exec(forumUpdateQuery)

		if err != nil {
			fmt.Print("forum Update threads num err:")
			fmt.Println(err)
		}

		tx.Commit()
		w.WriteHeader(http.StatusCreated) //201
		response, _ := json.Marshal(newThread)
		w.Write(response)

	} else if (err != nil) && (strings.Contains(err.Error(), "pq: duplicate key")) {
		tx.Rollback()
		fmt.Print(" err :")
		log.Println(err)
		w.WriteHeader(http.StatusConflict) //409
		var existingThread structs.Thread
		existingThread.Id = newThread.Id
		row := env.DB.QueryRow("SELECT * FROM thread WHERE id = '" + strconv.Itoa(existingThread.Id) + "'")
		scanErr1 := row.Scan(&existingThread.Title, &existingThread.Author, &existingThread.Forum, &existingThread.Message, &existingThread.Created)

		if scanErr1 != nil {
			fmt.Print("Scan err1:")
			log.Println(scanErr1)
		}

		response, _ := json.Marshal(existingThread)
		w.Write(response)

	} else if (err != nil) && (strings.Contains(err.Error(), "threadfk2")) {
		tx.Rollback()
		fmt.Print("Create thread threadfk2 err:")
		log.Println(err)

		var existingThread structs.Thread
		row := env.DB.QueryRow("SELECT id, title, author, forum, message, slug, created FROM thread WHERE slug = '" + newThread.Slug + "'")
		scanErr2 := row.Scan(&existingThread.Id, &existingThread.Title, &existingThread.Author, &existingThread.Forum, &existingThread.Message, &existingThread.Slug, &existingThread.Created)

		if scanErr2 != nil {
			fmt.Print("Create thread scan err2:")
			log.Println(scanErr2)
			w.WriteHeader(http.StatusNotFound) //404
			errorMsg := map[string]string{"message": "Can't find user with id " + newThread.Author + "\n"}
			response, _ := json.Marshal(errorMsg)
			w.Write(response)
			return
		}

		w.WriteHeader(http.StatusConflict) //409
		response, _ := json.Marshal(existingThread)
		w.Write(response)
	}
}

func (env *Env) GetThreadDetails(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	var currentThread structs.Thread
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil
	sqlStatement := ""

	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil {
		sqlStatement = `SELECT * FROM thread WHERE id = $1`
	} else {
		sqlStatement = `SELECT * FROM thread WHERE slug = $1`
	}
	row := env.DB.QueryRow(sqlStatement, vars["slug_or_id"])
	err = row.Scan(&currentThread.Id, &currentThread.Title, &currentThread.Author, &currentThread.Forum, &currentThread.Message, &currentThread.Votes, &currentThread.Slug, &currentThread.Created)

	if err == nil {
		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(currentThread)
		w.Write(response)
	} else {
		fmt.Print("\n GetThreadDetails: ")
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id " + currentThread.Author + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}

func (env *Env) UpdateThreadDatails(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	var currentThread structs.Thread
	numFieldsToUpdate := 0
	json.NewDecoder(r.Body).Decode(&currentThread) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil
	sqlStatement := ""
	returnFields := ""
	emptyThread := structs.Thread{}

	if currentThread == emptyThread {
		sqlStatement = "SELECT * FROM thread "
	} else {
		sqlStatement = "UPDATE thread SET "
		returnFields = " RETURNING * "

		if currentThread.Message != "" {
			numFieldsToUpdate++
			sqlStatement += "message = '" + currentThread.Message + "'"
		}

		if currentThread.Title != "" && numFieldsToUpdate > 0 {
			sqlStatement += ", title = '" + currentThread.Title + "'"
		} else if currentThread.Title != "" {
			sqlStatement += "title = '" + currentThread.Title + "'"
		}
	}

	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil {
		sqlStatement += " WHERE id = " + vars["slug_or_id"] + returnFields
	} else {
		sqlStatement += " WHERE slug = '" + vars["slug_or_id"] + "'" + returnFields
	}

	row := env.DB.QueryRow(sqlStatement)
	err = row.Scan(&currentThread.Id, &currentThread.Title, &currentThread.Author, &currentThread.Forum, &currentThread.Message, &currentThread.Votes, &currentThread.Slug, &currentThread.Created)

	if err == nil {
		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(currentThread)
		w.Write(response)
	} else {
		fmt.Print("\n UpdateThreadDetails: ")
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id " + currentThread.Author + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}

func (env *Env) GetThreadPosts(w http.ResponseWriter, r *http.Request) {
	var err error
	params := r.URL.Query()
	limit := params.Get("limit")
	since := params.Get("since")
	sort := params.Get("sort")
	desc := params.Get("desc")

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	vars := mux.Vars(r)
	var postsInThread []structs.Post
	var currentPost structs.Post
	identific := ""
	sqlStatement := ""
	curThreadId := 0

	//----------------------thread check---------------------------
	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil {
		sqlStatement = `SELECT id FROM thread WHERE id = $1`
	} else {
		sqlStatement = `SELECT id FROM thread WHERE slug = $1`
	}

	row := env.DB.QueryRow(sqlStatement, vars["slug_or_id"])
	scanErr := row.Scan(&curThreadId)

	if scanErr != nil {
		fmt.Print("\n getThreadPosts threadNotFound:")
		fmt.Print(scanErr)
		fmt.Print("\n")

		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find thread by slug_or_id " + vars["slug_or_id"] + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}
	//-------------------------------------------------------------

	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil {
		sqlStatement = `SELECT P.id, P.parent, P.author, P.message, P.forum, P.thread, P.created
		 FROM post P WHERE thread = $1`
		identific = `WHERE P.thread = $1`
	} else {
		sqlStatement = `SELECT P.id, P.parent, P.author, P.message, P.forum, P.thread, P.created
		FROM post P JOIN thread ON (thread = thread.Id)
		WHERE slug = $1`
		identific = `WHERE slug = $1`
	}

	if sort == "" {
		sort = "flat"
	}

	if sort == "flat" {

		if since != "" && desc == "true" {
			sqlStatement += "AND P.id < " + since
		} else if since != "" {
			sqlStatement += "AND P.id > " + since
		}

		if desc == "true" {
			//fmt.Print("sort flat, desc")
			sqlStatement += " order by P.Id DESC "
		} else {
			sqlStatement += " order by created, P.Id "
		}

	} else if sort == "tree" {

		if since != "" && desc != "true" {
			//sqlStatement += "AND path > (SELECT path from post where id = " + since + ") order by path"
			sqlStatement += " AND path > (SELECT path from post where id = " + since + ") "
		} else if since != "" && desc == "true" {
			sqlStatement += " AND path < (SELECT path from post where id = " + since + ") "
		}

		if desc == "true" {
			sqlStatement += " ORDER BY P.path DESC "
		} else {
			sqlStatement += " ORDER BY P.path "
		}

	} else if sort == "parent_tree" {

		if since != "" && desc != "true" {
			sqlStatement += " AND path[1] > (SELECT path[1] from post where id = " + since + ") "
		} else if since != "" && desc == "true" {
			sqlStatement += " AND path[1] < (SELECT path[1] from post where id = " + since + ") "
		}

		if limit != "" {
			parentsStr := ""
			getParentsQuery := "SELECT P.id FROM post P JOIN thread ON (thread = thread.Id)" + identific + ` and parent = 0`

			if desc == "true" {
				getParentsQuery += " ORDER BY P.id DESC "
			} else {
				getParentsQuery += " ORDER BY P.id "
			}

			if since == "" {
				getParentsQuery += ` LIMIT ` + limit
			}

			//fmt.Print("\n PARENT TREE getParentsQuery: ")
			//fmt.Print(getParentsQuery)

			rows, queryErr := env.DB.Query(getParentsQuery, vars["slug_or_id"])

			if queryErr != nil {
				fmt.Print("\n parent_tree err: ")
				fmt.Print(queryErr)
			}

			curParent := 0
			rowsCount := 0

			for rows.Next() {
				err = rows.Scan(&curParent)
				if err != nil {
					fmt.Print("\n curParent err:")
					fmt.Print(err)
				} else {
					if rowsCount != 0 {
						parentsStr += ", " + strconv.Itoa(curParent)
					} else {
						parentsStr += strconv.Itoa(curParent)
					}
				}
				rowsCount++
			}
			sqlStatement += " AND path && ARRAY[" + parentsStr + "] "
		}

		if desc == "true" {
			//sqlStatement += " ORDER BY P.path[1] DESC, P.id"
			sqlStatement += "ORDER BY P.path[1] DESC, P.path"
		} else {
			sqlStatement += " ORDER BY path "
		}

		//fmt.Print("\n PARENT TREE LIMIT : ")
		//fmt.Print(sqlStatement)
	}

	if limit != "" && sort != "parent_tree" {
		sqlStatement += " limit " + limit
		//fmt.Print("\n query: ")
		//fmt.Print(sqlStatement)
	}

	rows, queryErr := env.DB.Query(sqlStatement, vars["slug_or_id"])
	if queryErr != nil {
		fmt.Print("\n queryErr1: ")
		fmt.Print(queryErr)
	}

	rowsNum := 0
	for rows.Next() {
		err = rows.Scan(&currentPost.Id, &currentPost.Parent, &currentPost.Author, &currentPost.Message, &currentPost.Forum, &currentPost.Thread, &currentPost.Created)
		if err != nil {
			fmt.Print("\n rowNext err:")
			fmt.Print(err)
		}
		rowsNum++
		postsInThread = append(postsInThread, currentPost)
	}

	if postsInThread == nil {
		var emptyThread [0]structs.Thread
		response, _ := json.Marshal(emptyThread)
		w.Write(response)
	} else {
		response, _ := json.Marshal(postsInThread)
		w.Write(response)
	}
}
