package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/user/{nickname}/profile", userProfileHandler)
	router.HandleFunc("/api/user/{nickname}/create", createUser)
	router.HandleFunc("/api/forum/create", createForum)
	router.HandleFunc("/api/forum/{slug}/details", getForumDetails)
	router.HandleFunc("/api/forum/{slug}/create", createThread)
	router.HandleFunc("/api/forum/{slug}/threads", getThreads).Methods("GET")
	router.HandleFunc("/api/thread/{slug_or_id}/create", createPost)
	router.HandleFunc("/api/thread/{slug_or_id}/vote", createVote)
	router.HandleFunc("/api/thread/{slug_or_id}/details", getThreadDetails).Methods("GET")
	router.HandleFunc("/api/thread/{slug_or_id}/details", updateThreadDatails).Methods("POST")
	router.HandleFunc("/api/thread/{slug_or_id}/posts", getThreadPosts).Methods("GET")
	router.HandleFunc("/api/forum/{slug_or_id}/users", getForumUsers).Methods("GET")
	router.HandleFunc("/api/post/{id}/details", getPostDetails).Methods("GET")
	router.HandleFunc("/api/post/{id}/details", updatePostDetails).Methods("POST")
	router.HandleFunc("/api/service/status", serviceGetStatus).Methods("GET")
	router.HandleFunc("/api/service/clear", serviceClear).Methods("POST")

	// http.HandleFunc("/pages/",
	// 	func(w http.ResponseWriter, r *http.Request) {
	// 		fmt.Fprintln(w, "Multiple pages:", r.URL.String())
	// 	})
	//http.HandleFunc("/", mainHandler)

	fmt.Println("starting server at :5000")
	//http.ListenAndServe(":5000", nil)
	http.ListenAndServe(":5000", router)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func connectToDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	//defer db.Close()
	return db
}

func logger(r http.Request) {
	start := time.Now()
	log.Printf(
		"%s\t%s\t\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

func serviceGetStatus(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
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

func serviceClear(w http.ResponseWriter, r *http.Request) {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
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
