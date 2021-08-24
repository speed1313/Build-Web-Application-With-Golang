package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //urlが渡すオプションを解析します。POSTに対してはレスポンスパケットのボディを解析します（request body）
	//注意：もしParseFormメソッドがコールされなければ、以下でフォームのデータを取得することができません。
	fmt.Println(r.Form) //これらのデータはサーバのプリント情報に出力されます
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!") //ここでwに書き込まれたものがクライアントに出力されます。
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //リクエストを取得するメソッド
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("html/login.gtpl")
		t.Execute(w, token)
	} else {
		//ログインデータがリクエストされ、ログインのロジック判断が実行されます。
		r.ParseForm()
		token := r.Form.Get("token")
		if token != "" {
			//tokenの合法性を検証します。
			var cookie = getCookie(w, r, "token")
			if cookie == nil {
				setCookie(w, "token", token)
				fmt.Println("クッキーをセットしたよ！ : ", token)
			} else if cookie != nil && cookie.Value != token {
				fmt.Println("2回更新しているよ!")
				fmt.Println("送信された値", token)
				fmt.Println("保持している値", cookie.Value)
			}
		} else {
			fmt.Println("tokenがないです！")
		}
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
		fmt.Println("username:", template.HTMLEscapeString(r.Form.Get("username"))) //サーバ側に出力されます。
		fmt.Println("password:", template.HTMLEscapeString(r.Form.Get("password")))
		template.HTMLEscape(w, []byte(r.Form.Get("username"))) //クライアントに出力されます。
		t, err := template.New("foo").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
		err = t.ExecuteTemplate(w, "T", template.HTML("<script>alert('you have been pwned')</script>"))
		if err != nil {
			fmt.Println("error")
		}
		fmt.Println("age: ", r.Form["age"])
		fmt.Println("GET query string", r.URL)
		// バリデーションしてみる
		errors := validate(r.Form)
		fmt.Println(errors)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
    fmt.Println("method:", r.Method) //リクエストを受け取るメソッド
    if r.Method == "GET" {
        crutime := time.Now().Unix()
        h := md5.New()
        io.WriteString(h, strconv.FormatInt(crutime, 10))
        token := fmt.Sprintf("%x", h.Sum(nil))

        t, _ := template.ParseFiles("upload.html")
        t.Execute(w, token)
    } else {
        r.ParseMultipartForm(32 << 20)
        file, handler, err := r.FormFile("uploadfile")
        if err != nil {
            fmt.Println(err)
            return
        }
        defer file.Close()
        fmt.Fprintf(w, "%v", handler.Header)
        f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
        if err != nil {
            fmt.Println(err)
            return
        }
        defer f.Close()
        io.Copy(f, file)
    }
}

func main() {
	http.HandleFunc("/", sayhelloName) //アクセスのルーティングを設定します
	http.HandleFunc("/login", login)   //アクセスのルーティングを設定します
	http.HandleFunc("/upload", upload)
	err := http.ListenAndServe(":9090", nil) //監視するポートを設定します
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func validate(form url.Values) (errors []string) {
	const requiredErr = "%sは必須です。"
	const numErr = "%sは数字で入力してください。"
	const rangeErr = "%sは0～99の値で入力してください。"
	if len(form["username"][0]) == 0 {
		errors = append(errors, fmt.Sprintf(requiredErr, "ユーザ名"))
	}
	if form.Get("age") != "" {
		getint, err := strconv.Atoi(form.Get("age"))
		if err != nil {
			errors = append(errors, fmt.Sprintf(numErr, "年齢"))
		} else {
			if getint < 0 || getint > 99 {
				errors = append(errors, fmt.Sprintf(rangeErr, "年齢"))
			}
		}
	}
	return
}

func setCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}
	http.SetCookie(w, cookie)
}

func getCookie(w http.ResponseWriter, r *http.Request, name string) *http.Cookie {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil
	}
	return cookie
}