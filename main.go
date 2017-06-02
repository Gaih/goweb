package main

import (
	"flag"
	"gowiki/src/wiki"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	_ "github.com/astaxie/session/providers/memory"
	_ "github.com/go-sql-driver/mysql"
	"gowiki/src/user"
)

var (
	validPath = regexp.MustCompile("^/(edit|save|view|login|main)/([a-zA-Z0-9]+)$")
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)
func main() {
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(wiki.ViewHandler))
	http.HandleFunc("/edit/", makeHandler(wiki.EditHandler))
	http.HandleFunc("/save/", makeHandler(wiki.SaveHandler))
	http.HandleFunc("/login/", user.LoginHandler)
	http.HandleFunc("/main/", user.MainHandler)
	http.HandleFunc("/logout/", user.LogoutHandler)
	http.HandleFunc("/postmain/", user.PostHandler)
	http.HandleFunc("/change/", user.ChangeHandler)
	http.HandleFunc("/admin/", user.AdminHandler)
	http.HandleFunc("/userinfo/",user.UserInfoHandler)
	http.HandleFunc("/postselect/",user.SelectHandler)

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

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		log.Println("m=", m, "r:", r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}
