package main

import (
	"time"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "123"
	DB_NAME     = "Forum"
)

type User struct {
	Nickname string `json:"nickname"`
	FullName string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}

type Forum struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	User    string `json:"user"`
	Posts   int    `json:"posts"`
	Threads int    `json:"threads"`
}

// type ForumFullInfo struct {
// 	Slug    string `json:"slug"`
// 	Title   string `json:"title"`
// 	User    string `json:"user"`
// 	Posts   int    `json:"posts"`
// 	Threads int    `json:"threads"`
// }

type Thread struct {
	Id      int       `json:"id"`
	Title   string    `json:"title"`
	Author  string    `json:"author"`
	Forum   string    `json:"forum"`
	Message string    `json:"message"`
	Votes   int       `json:"votes,omitempty"`
	Slug    string    `json:"slug,omitempty"`
	Created time.Time `json:"created"`
}

type Post struct {
	Id       int       `json:"id"`
	Parent   int       `json:"parent,omitempty"`
	Author   string    `json:"author"`
	Message  string    `json:"message"`
	IsEdited bool      `json:"isEdited,omitempty"`
	Forum    string    `json:"forum"`
	Thread   int       `json:"thread"`
	Created  time.Time `json:"created"`
	Path     string    `json:"-"`
}

type PostsArray struct {
	Posts []Post `json:"posts"`
}

type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int    `json:"voice"`
	ThreadId int    `json:"-"`
}

type PostRelated struct {
	PostUser   *User   `json:"author,omitempty"`
	PostForum  *Forum  `json:"forum,omitempty"`
	PostThread *Thread `json:"thread,omitempty"`
	MainPost   *Post   `json:"post"`
}
