package main

import (
	"fmt"
	"github.com/astaxie/session"
	_ "github.com/astaxie/session/providers/memory"
	"html/template"
	"log"
	"net/http"
	"time"
)

var globalSessions *session.Manager
var loginSession session.Session

func main() {
	http.HandleFunc("/", index)              //アクセスのルーティングを設定します
	http.HandleFunc("/login", login)         //アクセスのルーティングを設定します
	http.HandleFunc("/count", count)         //アクセスのルーティングを設定します
	err := http.ListenAndServe(":9090", nil) //監視するポートを設定します
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
}

func login(w http.ResponseWriter, r *http.Request) {
	loginSession = globalSessions.SessionStart(w, r)
	r.ParseForm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/login.html")
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, loginSession.Get("username"))
	} else {
		loginSession.Set("username", r.Form["username"])
		http.Redirect(w, r, "/", 302)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if loginSession != nil {
		fmt.Println("sessionId: ", loginSession.SessionID)
		fmt.Println("Username: ", loginSession.Get("username"))
	} else {
		fmt.Println("loginSession is nil")
	}
	fmt.Fprintf(w, "Hello astaxie!") //ここでwに書き込まれたものがクライアントに出力されます。
}

func count(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	createtime := sess.Get("createtime")
	if createtime == nil {
		sess.Set("createtime", time.Now().Unix())
	} else if (createtime.(int64) + 360) < (time.Now().Unix()) {
		globalSessions.SessionDestroy(w, r)
		sess = globalSessions.SessionStart(w, r)
	}
	ct := sess.Get("countnum")
	if ct == nil {
		sess.Set("countnum", 1)
	} else {
		sess.Set("countnum", (ct.(int) + 1))
	}
	t, _ := template.ParseFiles("html/count.html")
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, sess.Get("countnum"))
}
