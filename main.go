package main

import (
	"flag"
	"gowiki/src/wiki"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

var (
	templates = template.Must(template.ParseFiles("html/user/login.html", "html/user/main.html"))
	validPath = regexp.MustCompile("^/(edit|save|view|login|main)/([a-zA-Z0-9]+)$")
	addr      = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

func main() {
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(wiki.ViewHandler))
	http.HandleFunc("/edit/", makeHandler(wiki.EditHandler))
	http.HandleFunc("/save/", makeHandler(wiki.SaveHandler))
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/main/", mainHandler)

	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}
	http.ListenAndServe(":8080", nil)
}
func loginHandler(writer http.ResponseWriter, request *http.Request) {
	err := templates.ExecuteTemplate(writer, "login.html", nil)
	log.Println("登陆界面")
	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
func mainHandler(writer http.ResponseWriter, request *http.Request) {
	email := request.FormValue("inputEmail")
	password := request.FormValue("inputPassword")

	err := templates.ExecuteTemplate(writer, "main.html", nil)
	log.Println("用户主界面", email+password)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		log.Println("m=", m, "r:", r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[1])
	}
}
