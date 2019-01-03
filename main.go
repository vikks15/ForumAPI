package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	//"ForumAPI/handlers"
	//"ForumAPI/structs"

	"github.com/vikks15/ForumAPI/handlers"
	"github.com/vikks15/ForumAPI/structs"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		structs.DB_HOST, structs.DB_PORT, structs.DB_USER, structs.DB_PASSWORD, structs.DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	handlersEnv := &handlers.Env{DB: db}

	router := mux.NewRouter()
	router.HandleFunc("/api/user/{nickname}/profile", handlersEnv.UserProfileHandler)
	router.HandleFunc("/api/user/{nickname}/create", handlersEnv.CreateUser)
	router.HandleFunc("/api/forum/create", handlersEnv.CreateForum)
	router.HandleFunc("/api/forum/{slug}/details", handlersEnv.GetForumDetails)
	router.HandleFunc("/api/forum/{slug}/create", handlersEnv.CreateThread)
	router.HandleFunc("/api/forum/{slug}/threads", handlersEnv.GetThreads).Methods("GET")
	router.HandleFunc("/api/thread/{slug_or_id}/create", handlersEnv.CreatePost)
	router.HandleFunc("/api/thread/{slug_or_id}/vote", handlersEnv.CreateVote)
	router.HandleFunc("/api/thread/{slug_or_id}/details", handlersEnv.GetThreadDetails).Methods("GET")
	router.HandleFunc("/api/thread/{slug_or_id}/details", handlersEnv.UpdateThreadDatails).Methods("POST")
	router.HandleFunc("/api/thread/{slug_or_id}/posts", handlersEnv.GetThreadPosts).Methods("GET")
	router.HandleFunc("/api/forum/{slug_or_id}/users", handlersEnv.GetForumUsers).Methods("GET")
	router.HandleFunc("/api/post/{id}/details", handlersEnv.GetPostDetails).Methods("GET")
	router.HandleFunc("/api/post/{id}/details", handlersEnv.UpdatePostDetails).Methods("POST")
	router.HandleFunc("/api/service/status", handlersEnv.ServiceGetStatus).Methods("GET")
	router.HandleFunc("/api/service/clear", handlersEnv.ServiceClear).Methods("POST")

	fmt.Println("starting server at :5000")
	http.ListenAndServe(":5000", router)
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
