package router

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sadnessOjisan/gochann/model"
)

// see: https://stackoverflow.com/questions/15130321/is-there-a-method-to-generate-a-uuid-with-go-language
func pseudoUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (h *Handler) UsersDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Printf("method not allowed")
		return
	}
	sub := strings.TrimPrefix(r.URL.Path, "/users")
	_, id := filepath.Split(sub)
	if id == "" {
		log.Printf("ERROR: user id is not found err")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	u := &model.User{}
	row := h.db.QueryRowContext(ctx, "select * from users where id = ? limit 1", id)
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("ERROR: db scan user err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Println(err)
	}
}

func (h *Handler) UsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// POST users
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		if !(utf8.RuneCountInString(name) >= 1 && utf8.RuneCountInString(name) <= 32) {
			log.Printf("ERROR: name length is not invalid name: %s, utf8.RuneCountInString(name): %d", name, utf8.RuneCountInString(name))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		exsistUserRow := h.db.QueryRowContext(ctx, "select id, password, salt from users where name = ? limit 1", name)
		var (
			userID                                     int
			currentUserSalt, currentUserHashedPassword string
		)
		if err := exsistUserRow.Scan(&userID, &currentUserHashedPassword, &currentUserSalt); err != nil && err != sql.ErrNoRows {
			log.Printf("ERROR: db scan user err: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		password := r.FormValue("password")
		if !(utf8.RuneCountInString(password) >= 1 && utf8.RuneCountInString(password) <= 100) {
			log.Printf("ERROR: title length is not invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if userID == 0 {
			// アカウント情報が存在しないなら登録してクッキーを発行する
			salt := pseudoUUID()
			passwordAddedSalt := password + salt
			passwordByte := []byte(passwordAddedSalt)
			hasher := sha256.New()
			hasher.Write(passwordByte)
			hashedPasswordString := hex.EncodeToString(hasher.Sum(nil))

			ins, err := h.db.PrepareContext(ctx, "insert into users(name, password, salt) value (?, ?, ?)")
			if err != nil {
				log.Printf("ERROR: prepare users insert err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := ins.ExecContext(ctx, name, hashedPasswordString, salt)
			if err != nil {
				log.Printf("ERROR: exec user insert err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			addedUserID, err := res.LastInsertId()
			if err != nil {
				log.Printf("ERROR: get last user id err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			uuid := pseudoUUID()

			sessionInsert, err := h.db.PrepareContext(ctx, "insert into session(user_id, token) value (?, ?)")
			if err != nil {
				log.Printf("ERROR: prepare session insert err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = sessionInsert.ExecContext(ctx, addedUserID, uuid)
			if err != nil {
				log.Printf("ERROR: exec session insert err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			cookie := &http.Cookie{
				Name:     "token",
				Value:    uuid,
				Expires:  time.Now().AddDate(0, 0, 1),
				SameSite: http.SameSiteStrictMode,
				HttpOnly: true,
				Secure:   true,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/posts", http.StatusSeeOther)
			return
		} else {
			// アカウント情報が存在するユーザーなら、入力されたパスワードと正しいか確認してから、クッキー発行してログインさせる
			passwordAddedSalt := password + currentUserSalt
			passwordByte := []byte(passwordAddedSalt)
			hasher := sha256.New()
			hasher.Write(passwordByte)
			hashedPasswordString := hex.EncodeToString(hasher.Sum(nil))

			if currentUserHashedPassword != hashedPasswordString {
				log.Printf("ERROR: user input password mismatch")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			uuid := pseudoUUID()
			sessionInsert, err := h.db.PrepareContext(ctx, "insert into session(user_id, token) value (?, ?)")
			if err != nil {
				log.Printf("ERROR: prepare session insert err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = sessionInsert.Exec(userID, uuid)
			if err != nil {
				log.Printf("ERROR: exec session insert err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			cookie := &http.Cookie{
				Name:     "token",
				Value:    uuid,
				Expires:  time.Now().AddDate(0, 0, 1),
				SameSite: http.SameSiteStrictMode,
				HttpOnly: true,
				Secure:   true,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/posts", http.StatusSeeOther)
			return
		}
	}

	// GET /posts
	if r.Method == http.MethodGet {
		rows, err := h.db.QueryContext(ctx, "select * from users")
		if err != nil {
			log.Printf("ERROR: exec users query err: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var users []*model.User
		for rows.Next() {
			u := &model.User{}
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
				log.Printf("ERROR: db scan users err: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(users); err != nil {
			log.Println(err)
		}
	}
}
