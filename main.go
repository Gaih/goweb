package main

import (
	"database/sql"
	"flag"
	"gowiki/src/wiki"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"

	"fmt"
	"time"

	"github.com/astaxie/session"
	_ "github.com/astaxie/session/providers/memory"
	_ "github.com/go-sql-driver/mysql"
)

var (
	templates = template.Must(template.ParseFiles("html/user/login.html", "html/user/main.html"))
	validPath = regexp.MustCompile("^/(edit|save|view|login|main)/([a-zA-Z0-9]+)$")
	addr      = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

func main() {

	// db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	// //插入数据
	// defer db.Close()
	// stmt, err := db.Prepare("INSERT userdata SET uid=?,name=?,new_num=?,tol_num=?,id=?,date=?")
	// checkErr(err)

	// res, err := stmt.Exec("1", "蔚蓝少女", "12", "34", "1", time.Now().Format("2006-01-02"))
	// checkErr(err)

	// id, err := res.RowsAffected()
	// checkErr(err)

	// fmt.Println(id)

	// //查询数据
	// rows, err := db.Query("SELECT * FROM userdata")
	// checkErr(err)
	// for rows.Next() {
	// 	var uid int
	// 	var name string
	// 	var newnum string
	// 	var tolnum string
	// 	var id int
	// 	var date string
	// 	err = rows.Scan(&uid, &name, &newnum, &tolnum, &id, &date)
	// 	checkErr(err)
	// 	fmt.Println(uid)
	// 	fmt.Println(name)
	// 	fmt.Println(newnum)
	// 	fmt.Println(tolnum)
	// 	fmt.Println(date)
	// }
	// //构造scanArgs、values两个数组，scanArgs的每个值指向values相应值的地址
	// 	columns, _ := rows.Columns()
	// 	scanArgs := make([]interface{}, len(columns))
	// 	values := make([]interface{}, len(columns))
	// 	for i := range values {
	// 		scanArgs[i] = &values[i]
	// 	}

	// 	for rows.Next() {
	// 		//将行数据保存到record字典
	// 		err = rows.Scan(scanArgs...)
	// 		record := make(map[string]string)
	// 		for i, col := range values {
	// 			if col != nil {
	// 				record[columns[i]] = string(col.([]byte))
	// 			}
	// 		}
	// 		fmt.Println(record)
	// 	}

	flag.Parse()
	http.HandleFunc("/view/", makeHandler(wiki.ViewHandler))
	http.HandleFunc("/edit/", makeHandler(wiki.EditHandler))
	http.HandleFunc("/save/", makeHandler(wiki.SaveHandler))
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/main/", mainHandler)
	http.HandleFunc("/logout/", logoutHandler)

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
func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	globalSessions.SessionDestroy(writer, request)
	http.Redirect(writer, request, "/login/", http.StatusFound)
	return
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
	//globalSessions.SessionDestroy(writer, request)

	if request.Method == "POST" {
		email := request.FormValue("inputEmail")
		inputPassword := request.FormValue("inputPassword")

		db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
		defer db.Close()
		// fmt.Printf("SELECT password FROM userinfo where account=\"%v\"", email)
		//查询数据
		res, err := db.Prepare("SELECT * FROM userinfo where account=?")

		rows, err := res.Query(email)
		checkErr(err)
		var id int
		var username string
		var account string
		var password string
		var admin int
		for rows.Next() {
			err = rows.Scan(&id, &username, &account, &password, &admin)
			checkErr(err)
		}

		//密码正确 跳转到主界面
		if inputPassword == password {
			sess := globalSessions.SessionStart(writer, request)

			log.Println(sess.SessionID(), "输入密码", inputPassword, "密码：", password, "id:", id, "用户名：", username, "管理员：", admin)

			http.Redirect(writer, request, "/main/", http.StatusFound)
			//ct := sess.Get("username")
			//if ct ==nil {
			sess.Set("id", id)
			sess.Set("username", username)
			//}
			return
		} else {
			http.Redirect(writer, request, "/login/", http.StatusFound)
			return
		}
	} else {
		log.Println("重新登陆")
		err := templates.ExecuteTemplate(writer, "login.html", nil)
		log.Println("登陆界面")
		if err != nil {
			log.Println(err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

	}

}
func mainHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")
	createtime := sess.Get("createtime")
	if createtime == nil {
		sess.Set("createtime", time.Now().Unix())
	} else if (createtime.(int64) + 60*60*24) < (time.Now().Unix()) {
		globalSessions.SessionDestroy(writer, request)
		sess = globalSessions.SessionStart(writer, request)
	}
	log.Println("URL", request.URL.Path, "SessionID:", sess.SessionID(), "username:", ct)
	if request.URL.Path == "/main/" {
		if ct != nil {
			//m := validPath.FindStringSubmatch(request.URL.Path)
			//log.Println("m=", m)
			//if m == nil{
			//	http.NotFound(writer, request)
			//	return
			//}
			//p := new(Page)
			//p.Username = username
			//p.Account = account

			//s数据库查询账户数据
			uid := sess.Get("id")
			db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
			defer db.Close()
			// fmt.Printf("SELECT password FROM userinfo where account=\"%v\"", email)
			//查询数据
			res, err := db.Prepare("SELECT * FROM userdata where uid=?")

			query, err := res.Query(uid)
			checkErr(err)

			column, _ := query.Columns()              //读出查询出的列字段名
			values := make([][]byte, len(column))     //values是每个列的值，这里获取到byte里
			scans := make([]interface{}, len(column)) //因为每次查询出来的列是不定长的，用len(column)定住当次查询的长度
			for i := range values {
				//让每一行数据都填充到[][]byte里面
				scans[i] = &values[i]
			}
			results := make(map[int]map[string]string) //最后得到的map
			i := 0
			for query.Next() { //循环，让游标往下移动
				if err := query.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
					fmt.Println(err)
					return
				}
				row := make(map[string]string) //每行数据
				for k, v := range values { //每行数据是放在values里面，现在把它挪到row里
					// log.Println("value:", string(v))
					key := column[k]
					row[key] = string(v)
				}
				results[i] = row //装入结果集中
				i++
			}
			for k, v := range results { //查询出来的数组
				fmt.Println(k, v)
			}

			//for rows.Next() {
			//	err = rows.Scan(&GameData.id, &gamename, &new_num, &tol_num, &gameid, &date)
			//	log.Println("id:", id, "gamename:", gamename, "gameid:", gameid, "num:", new_num, tol_num, "time:", date)
			//	checkErr(err)
			//}
			value, ok := ct.(string)
			if ok {
				p := &GameData{Gamename:value, Data: results}
				log.Println("r:", request.URL.Path, "当前用户：", ct)
				er := templates.ExecuteTemplate(writer, "main.html", p)
				if er != nil {
					http.Error(writer, er.Error(), http.StatusInternalServerError)
				}
			}
			return
		} else {
			http.Redirect(writer, request, "/login/", http.StatusFound)
			return
		}
	} else {
		http.NotFound(writer, request)
		return
	}

}

type GameData struct {
	Gamename string
	Data     map[int]map[string]string
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

//session管理
var globalSessions *session.Manager

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}
