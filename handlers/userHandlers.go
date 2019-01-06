package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vikks15/ForumAPI/structs"
)

func (env *Env) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	//nickname := strings.TrimSuffix(strings.TrimPrefix(r.URL.String(), "/user/"), "/profile")
	var printUser structs.User
	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	switch r.Method {
	case "GET":
		row := env.DB.QueryRow("SELECT * FROM forumUser WHERE nickname = '" + vars["nickname"] + "'")
		var currentUser structs.User
		err = row.Scan(&currentUser.Nickname, &currentUser.FullName, &currentUser.About, &currentUser.Email)
		printUser = currentUser
	case "POST":
		var userDataToUpdate structs.User
		var updatedUser structs.User
		json.NewDecoder(r.Body).Decode(&userDataToUpdate)
		r.Body.Close()

		userDataToUpdate.Nickname = vars["nickname"]
		countFields := 0
		var res sql.Result
		var execErr error

		sqlStatement := "UPDATE forumUser SET "
		if userDataToUpdate.FullName != "" {
			sqlStatement += "fullname = '" + userDataToUpdate.FullName + "'"
			countFields++
		}
		if userDataToUpdate.About != "" {
			if countFields != 0 {
				sqlStatement += ", "
			}
			sqlStatement += "about = '" + userDataToUpdate.About + "'"
			countFields++
		}
		if userDataToUpdate.Email != "" {
			if countFields != 0 {
				sqlStatement += ", "
			}
			sqlStatement += "email = '" + userDataToUpdate.Email + "'"
			countFields++
		}
		if countFields != 0 {
			sqlStatement += " WHERE nickname = '" + userDataToUpdate.Nickname + "'"
			res, execErr = env.DB.Exec(sqlStatement)
		} else {
			sqlStatement = ""
		}

		var errExample = errors.New("pq: duplicate key value violates unique constraint \"unique_email\"")

		if execErr != nil && execErr.Error() == errExample.Error() {
			var userWithSameNick string
			w.WriteHeader(http.StatusConflict)
			//sameEmailUser := db.Model(User).Column("nickname").Where("email = ?", updatedUser.Email).Select()
			sameEmailUser := env.DB.QueryRow("SELECT nickname FROM forumUser WHERE email = '" + printUser.Email + "'")
			err = sameEmailUser.Scan(&userWithSameNick)
			errorMsg := map[string]string{"This email is already registered by user": userWithSameNick}
			response, _ := json.Marshal(errorMsg)
			w.Write(response)
			return
		}

		var rowsAf int64 = -1
		if res != nil {
			rowsAf, _ = res.RowsAffected()
		}

		if rowsAf == 0 {
			var errNoUser = errors.New("no User")
			err = errNoUser
		} else {
			row := env.DB.QueryRow("SELECT * FROM forumUser WHERE nickname = '" + vars["nickname"] + "'")
			err = row.Scan(&updatedUser.Nickname, &updatedUser.FullName, &updatedUser.About, &updatedUser.Email)

			if err == nil {
				response, _ := json.Marshal(updatedUser)
				w.Write(response)
				return
			}
			printUser = updatedUser
			err = execErr
		}
	}

	if err == nil {
		response, _ := json.Marshal(printUser)
		w.Write(response)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorMsg := map[string]string{"message": "Can't find user by nickname: " + vars["nickname"]}
		response, _ := json.Marshal(errorMsg)
		w.Write(response)
	}
}

func (env *Env) CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var newUser structs.User
	newUser.Nickname = vars["nickname"]
	json.NewDecoder(r.Body).Decode(&newUser) //request json to struct User
	r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header()["Date"] = nil

	tx, err := env.DB.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tx.Rollback()

	// _, err = tx.Exec("SET LOCAL synchronous_commit = OFF")

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	sqlStatement := `INSERT INTO forumUser (nickname, fullname, about, email) VALUES ($1,$2,$3,$4)`
	_, err = tx.Exec(sqlStatement, newUser.Nickname, newUser.FullName, newUser.About, newUser.Email)

	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		w.WriteHeader(http.StatusConflict)

		var existingUser1, existingUser2 structs.User
		row1 := env.DB.QueryRow("SELECT * FROM forumUser WHERE email = '" + newUser.Email + "'")
		row2 := env.DB.QueryRow("SELECT * FROM forumUser WHERE nickname = '" + newUser.Nickname + "'")
		_ = row1.Scan(&existingUser1.Nickname, &existingUser1.FullName, &existingUser1.About, &existingUser1.Email)
		_ = row2.Scan(&existingUser2.Nickname, &existingUser2.FullName, &existingUser2.About, &existingUser2.Email)
		var arr []structs.User

		if (existingUser1 == (structs.User{})) || (existingUser1 == existingUser2) {
			arr = []structs.User{
				existingUser2,
			}
		} else if existingUser2 == (structs.User{}) {
			arr = []structs.User{
				existingUser1,
			}
		} else {
			arr = []structs.User{
				existingUser1,
				existingUser2,
			}
		}

		result, _ := json.Marshal(arr)
		w.Write(result)
	} else {
		tx.Commit()
		result, _ := json.Marshal(newUser)
		w.WriteHeader(http.StatusCreated)
		w.Write(result)
	}
}
