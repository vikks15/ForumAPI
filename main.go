package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"ForumAPI/handlers"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/user/{nickname}/profile", handlers.UserProfileHandler)
	router.HandleFunc("/api/user/{nickname}/create", handlers.CreateUser)
	router.HandleFunc("/api/forum/create", handlers.CreateForum)
	router.HandleFunc("/api/forum/{slug}/details", handlers.GetForumDetails)
	router.HandleFunc("/api/forum/{slug}/create", handlers.CreateThread)
	router.HandleFunc("/api/forum/{slug}/threads", handlers.GetThreads).Methods("GET")
	router.HandleFunc("/api/thread/{slug_or_id}/create", handlers.CreatePost)
	router.HandleFunc("/api/thread/{slug_or_id}/vote", handlers.CreateVote)
	router.HandleFunc("/api/thread/{slug_or_id}/details", handlers.GetThreadDetails).Methods("GET")
	router.HandleFunc("/api/thread/{slug_or_id}/details", handlers.UpdateThreadDatails).Methods("POST")
	router.HandleFunc("/api/thread/{slug_or_id}/posts", handlers.GetThreadPosts).Methods("GET")
	router.HandleFunc("/api/forum/{slug_or_id}/users", handlers.GetForumUsers).Methods("GET")
	router.HandleFunc("/api/post/{id}/details", handlers.GetPostDetails).Methods("GET")
	router.HandleFunc("/api/post/{id}/details", handlers.UpdatePostDetails).Methods("POST")
	router.HandleFunc("/api/service/status", handlers.ServiceGetStatus).Methods("GET")
	router.HandleFunc("/api/service/clear", handlers.ServiceClear).Methods("POST")

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
