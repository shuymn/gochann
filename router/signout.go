package router

import (
	"log"
	"net/http"
	"time"
)

// どんな結果だろうと必ずクッキーを消すようにする。early return しない。
func (h *Handler) SignoutHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	ins, err := h.db.Prepare("delete from session where token =?")
	if err != nil {
		log.Printf("ERROR: prepare token delete err: %v", err)
	}
	_, err = ins.Exec(token.Value)
	if err != nil {
		log.Printf("ERROR: exec token delete err: %v", err)
	}

	cookie := &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
