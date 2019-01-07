package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/vikks15/ForumAPI/structs"
)

func (env *Env) CreatePost(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now()
	vars := mux.Vars(r)
	var newPosts []structs.Post
	var addedPosts []structs.Post
	forumUpdateQuery := ""
	numOfPosts := 0
	sqlStatement := ""

	err := json.NewDecoder(r.Body).Decode(&newPosts) //request json to struct User
	r.Body.Close()
	if err != nil {
		fmt.Println(err)
	}

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

	curPostThread := 0
	curPostForum := ""

	if _, err := strconv.Atoi(vars["slug_or_id"]); err == nil { //if id
		sqlStatement = "SELECT id, forum FROM thread where id = " + vars["slug_or_id"]
	} else {
		sqlStatement = "SELECT id, forum FROM thread WHERE slug = '" + vars["slug_or_id"] + "'"
	}
	row := tx.QueryRow(sqlStatement)
	scanErr := row.Scan(&curPostThread, &curPostForum)

	if scanErr != nil {
		fmt.Print("createPost NoThread err :")
		log.Println(scanErr)

		w.WriteHeader(http.StatusNotFound)
		//errorMsg := map[string]string{"message": "Can't find post thread by id: " + vars["slug_or_id"]}
		errorMsg := map[string]string{"message": "Can't find thread with slug or id " + vars["slug_or_id"] + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
		return
	}

	for _, post := range newPosts {
		if numOfPosts == 0 {
			currentTime = time.Now().UTC()
		}
		post.Created = currentTime.Round(time.Millisecond)

		post.Thread = curPostThread
		post.Forum = curPostForum
		parentThread := 0
		previousPath := ""

		//---------------------User case check--------------
		sqlStatement = `SELECT nickname FROM forumUser WHERE nickname = $1`
		row := tx.QueryRow("SELECT nickname FROM forumUser WHERE nickname = '" + post.Author + "'")
		scanErr := row.Scan(&post.Author)

		if scanErr != nil {
			fmt.Print("User case check: ")
			log.Println(scanErr)

			w.WriteHeader(http.StatusNotFound)
			errorMsg := map[string]string{"message": "Can't find user " + post.Author + "\n"}
			response, _ := json.Marshal(errorMsg)
			w.Write(response)
			return
		}
		//-----------------------------------preInsert to get post id--------------------------------
		sqlStatement = `INSERT INTO post (author, message, thread, forum, created, parent)
						VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
		preInsertRow := tx.QueryRow(sqlStatement, post.Author, post.Message, post.Thread, post.Forum, post.Created, post.Parent)
		err = preInsertRow.Scan(&post.Id)

		if err != nil {
			fmt.Print("\n Create post err in preInsert:")
			fmt.Println(err)
			return
		}
		//-------------------------------------------------------------------------------------------
		if strconv.Itoa(post.Parent) != "0" {
			sqlStatement = "SELECT thread, path FROM post where id = " + strconv.Itoa(post.Parent)
			row := tx.QueryRow(sqlStatement)
			err = row.Scan(&parentThread, &previousPath)

			if err != nil {
				fmt.Print("\n err in create new post with parent:")
				fmt.Print(err)
			}

			if parentThread != post.Thread {
				fmt.Print(err)
				w.WriteHeader(http.StatusConflict)
				errorMsg := map[string]string{"message": "Parent post was created in another thread"}
				response, _ := json.Marshal(errorMsg)
				w.Write(response)
				return
			}
			post.Path = strings.TrimRight(previousPath, "}") + "," + strconv.Itoa(post.Id) + "}"
		} else {
			post.Path = "{" + strconv.Itoa(post.Id) + "}"
		}

		sqlStatement = `UPDATE post SET path = $1 WHERE id = $2`
		//_, err = db.Exec(sqlStatement, post.Author, post.Message, post.Thread, post.Forum, post.Created, post.Parent)
		_, err = tx.Exec(sqlStatement, post.Path, post.Id)

		if err != nil {
			fmt.Print("\n Create post update path err:")
			fmt.Print(err)
			return
		}

		addedPosts = append(addedPosts, post)
		numOfPosts++
	}

	if numOfPosts == 0 {
		tx.Commit()
		w.WriteHeader(http.StatusCreated)
		var emptyPostArr [0]structs.Post
		response, _ := json.Marshal(emptyPostArr)
		w.Write(response)
		return
	} else {
		forumUpdateQuery = `UPDATE forum SET posts = posts + $1 WHERE slug = $2`
		_, err = tx.Exec(forumUpdateQuery, numOfPosts, curPostForum)

		if err != nil {
			fmt.Print("forum Update posts num err:")
			fmt.Println(err)
			return
		}

		tx.Commit()
		w.WriteHeader(http.StatusCreated)
		response, _ := json.Marshal(addedPosts)
		w.Write(response)
	}
}

func (env *Env) GetPostDetails(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	var currentPost structs.Post
	params := r.URL.Query()
	related := params.Get("related")

	json.NewDecoder(r.Body).Decode(&currentPost) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	sqlStatement := "SELECT * FROM post WHERE id = " + vars["id"]
	row := env.DB.QueryRow(sqlStatement)
	err = row.Scan(&currentPost.Id, &currentPost.Parent, &currentPost.Author, &currentPost.Message, &currentPost.IsEdited, &currentPost.Forum, &currentPost.Thread, &currentPost.Created, &currentPost.Path)

	if err == nil {
		w.WriteHeader(http.StatusOK)
		var response []byte
		var postRelated structs.PostRelated

		if strings.Contains(related, "user") {
			var postAuthor structs.User
			row = env.DB.QueryRow("SELECT * FROM forumUser WHERE nickname = '" + currentPost.Author + "'")
			userScanErr := row.Scan(&postAuthor.Nickname, &postAuthor.FullName, &postAuthor.About, &postAuthor.Email)

			if userScanErr != nil {
				fmt.Print("\n getPostPostDetails userScanErr: ")
				fmt.Print(userScanErr)
			}
			postRelated.PostUser = &postAuthor
		}

		if strings.Contains(related, "forum") {
			var postForum structs.Forum
			row = env.DB.QueryRow("SELECT * FROM forum WHERE slug = '" + currentPost.Forum + "'")
			forumScanErr := row.Scan(&postForum.Slug, &postForum.Title, &postForum.User, &postForum.Posts, &postForum.Threads)

			if forumScanErr != nil {
				fmt.Print("\n getPostPostDetails forumScanErr: ")
				fmt.Print(forumScanErr)
			}
			postRelated.PostForum = &postForum
		}

		if strings.Contains(related, "thread") {
			var postThread structs.Thread
			row = env.DB.QueryRow(`SELECT * FROM thread WHERE id = $1`, currentPost.Thread)
			threadScanErr := row.Scan(&postThread.Id, &postThread.Title, &postThread.Author, &postThread.Forum, &postThread.Message, &postThread.Votes, &postThread.Slug, &postThread.Created)

			if threadScanErr != nil {
				fmt.Print("\n getPostPostDetails threadScanErr: ")
				fmt.Print(threadScanErr)
			}
			postRelated.PostThread = &postThread
		}
		postRelated.MainPost = &currentPost
		response, _ = json.Marshal(postRelated)
		// responsePrep := map[string]Post{"post": currentPost}
		// response, _ = json.Marshal(responsePrep)
		w.Write(response)
	} else {
		fmt.Print("\n\n\n\n\n getPostPostDetails: ")
		fmt.Print(err)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id " + currentPost.Author + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}

func (env *Env) UpdatePostDetails(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	var postWithUpdate structs.Post
	var prevPost structs.Post
	var currentPost structs.Post

	json.NewDecoder(r.Body).Decode(&postWithUpdate) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	sqlStatement := "SELECT * FROM post WHERE id = " + vars["id"]
	row := env.DB.QueryRow(sqlStatement)
	err = row.Scan(&prevPost.Id, &prevPost.Parent, &prevPost.Author, &prevPost.Message, &prevPost.IsEdited, &prevPost.Forum, &prevPost.Thread, &prevPost.Created, &prevPost.Path)

	if err != nil {
		fmt.Print("\n getPost err in updatePostDetails: ")
		fmt.Print(err)
	}

	if prevPost.Message != postWithUpdate.Message && postWithUpdate.Message != "" {
		sqlUpdateStatement := "UPDATE post SET message = '" + postWithUpdate.Message + "', isedited = true WHERE id = " + vars["id"] + " RETURNING *"
		row = env.DB.QueryRow(sqlUpdateStatement)
		err = row.Scan(&currentPost.Id, &currentPost.Parent, &currentPost.Author, &currentPost.Message, &currentPost.IsEdited, &currentPost.Forum, &currentPost.Thread, &currentPost.Created, &currentPost.Path)
	} else {
		currentPost = prevPost
	}

	if err == nil {
		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(currentPost)
		w.Write(response)
	} else {
		fmt.Print("\n getPostPostDetails: ")
		fmt.Print(err)
		fmt.Print("\n")
		fmt.Print(sqlStatement)
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user with id " + currentPost.Author + "\n"}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}
