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

	"encoding/json"
	"io"

	"github.com/astaxie/session"
	_ "github.com/astaxie/session/providers/memory"
	_ "github.com/go-sql-driver/mysql"
)

var (
	templates = template.Must(template.ParseFiles("html/user/mainAdmin.html","html/user/login.html","html/user/password.html","html/user/main.html"))
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
	http.HandleFunc("/logout/", logoutHandler)
	http.HandleFunc("/postmain/", postHandler)
	http.HandleFunc("/change/", changeHandler)
	http.HandleFunc("/admin/", adminHandler)

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
func adminHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")
	// createtime := sess.Get("createtime")
	// if createtime == nil {
	// 	sess.Set("createtime", time.Now().Unix())
	// } else if (createtime.(int64) + 60*60*24) < (time.Now().Unix()) {
	// 	globalSessions.SessionDestroy(writer, request)
	// 	sess = globalSessions.SessionStart(writer, request)
	// }
	log.Println("URL:", request.URL.Path, "SessionID:", sess.SessionID(), "username:", ct)
	if request.URL.Path == "/admin/" {
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
			res, err := db.Prepare("SELECT date,name,new_num,tol_num FROM userdata where uid=? AND id=?")

			query, err := res.Query(uid, "1")
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
				for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
					// log.Println("value:", string(v))
					key := column[k]
					row[key] = string(v)
				}
				results[i] = row //装入结果集中
				i++
			}
			//m := 0
			//for _, v := range results { //查询出来的数组
			//	//log.Println(k, v)
			//	str, err := json.Marshal(v)
			//	if err == nil {
			//		log.Println(string(str))
			//	}
			//	jsonres[m] = string(str)
			//	m++
			//}
			log.Println(results)

			value, ok := ct.(string)
			if ok {
				p := &GameData{Gamename: value, Data: results}
				log.Println("r:", request.URL.Path, "当前用户：", ct, "p:", p)
				er := templates.ExecuteTemplate(writer, "mainAdmin.html", p)
				if er != nil {
					log.Println(er)
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
func changeHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method=="GET" {
		err := templates.ExecuteTemplate(writer, "password.html", nil)
		log.Println("修改密码")
		if err != nil {
			log.Println(err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}else if request.Method=="POST" {
		pass := request.FormValue("password")
		newPass := request.FormValue("reP")
		log.Println("password:",pass,"newPassword:",newPass)
		sess := globalSessions.SessionStart(writer, request)
		userId := sess.Get("id")

		db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
		defer db.Close()
		// fmt.Printf("SELECT password FROM userinfo where account=\"%v\"", email)
		//查询数据
		res, err := db.Prepare("SELECT password FROM userinfo where uid=?")

		rows, err := res.Query(userId)
		checkErr(err)
		var password string
		for rows.Next() {
			err = rows.Scan(&password)
			checkErr(err)
		}
		if pass == password {
			stmt, _ := db.Prepare("update userinfo set password=? where uid=?")
			result, err := stmt.Exec(string(newPass), userId)
			if err == nil {
				affect, err := result.RowsAffected()
				checkErr(err)
				log.Println("修改成功",affect)
				io.WriteString(writer, "修改成功")
				//http.Redirect(writer, request, "/main/", http.StatusFound)
			}
		}

	}
}

func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	globalSessions.SessionDestroy(writer, request)
	http.Redirect(writer, request, "/login/", http.StatusFound)
	return
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

			if admin == 0 {
				http.Redirect(writer, request, "/main/", http.StatusFound)
				//ct := sess.Get("username")
				//if ct ==nil {
				sess.Set("id", id)
				sess.Set("username", username)
			}else {
				http.Redirect(writer, request, "/admin/", http.StatusFound)
				//ct := sess.Get("username")
				//if ct ==nil {
				sess.Set("id", id)
				sess.Set("username", username)
			}

			//}
			return
		} else {
			http.Redirect(writer, request, "/login/", http.StatusFound)
			return
		}
	} else {
		err := templates.ExecuteTemplate(writer, "login.html", nil)
		log.Println("登陆界面")
		if err != nil {
			log.Println(err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

	}

}
func postHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")

	startTime := request.FormValue("start")
	endTime := request.FormValue("end")
	log.Println("starttime:", startTime, "endtime:", endTime)

	if startTime != ""{
		id := request.FormValue("nameid")
		uid := sess.Get("id")
		db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
		defer db.Close()
		// fmt.Printf("SELECT password FROM userinfo where account=\"%v\"", email)
		//查询数据
		res, err := db.Prepare("select  date,name,new_num,tol_num from userdata where date>=? and date<=? AND id=? AND uid=?")

		query, err := res.Query(startTime, endTime, id, uid)
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
			for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
				// log.Println("value:", string(v))
				key := column[k]
				row[key] = string(v)
			}
			results[i] = row //装入结果集中
			i++
		}
		//m := 0
		//for _, v := range results { //查询出来的数组
		//	//log.Println(k, v)
		//	str, err := json.Marshal(v)
		//	if err == nil {
		//		log.Println(string(str))
		//	}
		//	jsonres[m] = string(str)
		//	m++
		//}
		//log.Println("results:", results)

		value, ok := ct.(string)
		if ok {
			p := &GameData{Gamename: value, Data: results}
			log.Println("r:", request.URL.Path, "当前用户：", ct, "p:", p)
			b, err := json.Marshal(p)
			if err == nil {
				io.WriteString(writer, string(b))
			}
			//er := templates.ExecuteTemplate(writer, "main.html", p)
			//if er != nil {
			//	log.Println(er)
			//	http.Error(writer, er.Error(), http.StatusInternalServerError)
			//}
		}
	} else {
		id := request.FormValue("nameid")
		uid := sess.Get("id")
		db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
		defer db.Close()
		// fmt.Printf("SELECT password FROM userinfo where account=\"%v\"", email)
		//查询数据
		res, err := db.Prepare("SELECT date,name,new_num,tol_num FROM userdata where uid=? AND id=?")
		query, err := res.Query(uid, id)
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
			for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
				// log.Println("value:", string(v))
				key := column[k]
				row[key] = string(v)
			}
			results[i] = row //装入结果集中
			i++
		}
		//log.Println(results)

		value, ok := ct.(string)
		if ok {
			p := &GameData{Gamename: value, Data: results}
			log.Println("r:", request.URL.Path, "当前用户：", ct, "p:", p)
			b, err := json.Marshal(p)
			if err == nil {
				io.WriteString(writer, string(b))
			}
			//er := templates.ExecuteTemplate(writer, "main.html", p)
			//if er != nil {
			//	log.Println(er)
			//	http.Error(writer, er.Error(), http.StatusInternalServerError)
			//}
		}
	}
	return
}
func mainHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")
	// createtime := sess.Get("createtime")
	// if createtime == nil {
	// 	sess.Set("createtime", time.Now().Unix())
	// } else if (createtime.(int64) + 60*60*24) < (time.Now().Unix()) {
	// 	globalSessions.SessionDestroy(writer, request)
	// 	sess = globalSessions.SessionStart(writer, request)
	// }
	log.Println("URL:", request.URL.Path, "SessionID:", sess.SessionID(), "username:", ct)
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
			res, err := db.Prepare("SELECT date,name,new_num,tol_num FROM userdata where uid=? AND id=?")

			query, err := res.Query(uid, "1")
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
				for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
					// log.Println("value:", string(v))
					key := column[k]
					row[key] = string(v)
				}
				results[i] = row //装入结果集中
				i++
			}
			//m := 0
			//for _, v := range results { //查询出来的数组
			//	//log.Println(k, v)
			//	str, err := json.Marshal(v)
			//	if err == nil {
			//		log.Println(string(str))
			//	}
			//	jsonres[m] = string(str)
			//	m++
			//}
			log.Println(results)

			value, ok := ct.(string)
			if ok {
				p := &GameData{Gamename: value, Data: results}
				log.Println("r:", request.URL.Path, "当前用户：", ct, "p:", p)
				er := templates.ExecuteTemplate(writer, "main.html", p)
				if er != nil {
					log.Println(er)
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

type GameData struct {
	Gamename string
	Data     map[int]map[string]string
}
func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

//session管理
var globalSessions *session.Manager

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}
