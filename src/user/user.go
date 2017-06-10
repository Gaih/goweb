package user

import (
	"net/http"
	"log"
	"database/sql"
	"fmt"
	"io"
	"encoding/json"
	"html/template"
	"github.com/astaxie/session"
)

var templates = template.Must(template.ParseFiles("html/user/mainAdmin.html", "html/user/login.html", "html/user/password.html", "html/user/main.html"))
//session管理
var globalSessions *session.Manager

func AdminHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")
	log.Println("URL:", request.URL.Path, "SessionID:", sess.SessionID(), "username:", ct)
	if request.URL.Path == "/admin/" {
		if ct != nil {
			//s数据库查询账户数据
			//uid := sess.Get("id")
			db, _ := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
			defer db.Close()
			//查询数据
			rows, _ := db.Query("SELECT uid,username FROM userinfo ")

			var results map[int]string
			results = make(map[int]string)
			for rows.Next() {
				var uid int
				var username string
				err := rows.Scan(&uid, &username)
				checkErr(err)
				//fmt.Println(uid)
				//fmt.Println(username)
				results[uid] = username
			}

			//column, _ := res.Columns()              //读出查询出的列字段名
			//values := make([][]byte, len(column))     //values是每个列的值，这里获取到byte里
			//scans := make([]interface{}, len(column)) //因为每次查询出来的列是不定长的，用len(column)定住当次查询的长度
			//for i := range values {
			//	//让每一行数据都填充到[][]byte里面
			//	scans[i] = &values[i]
			//}
			//results := make(map[int]map[string]string) //最后得到的map
			//i := 0
			//for res.Next() { //循环，让游标往下移动
			//	if err := res.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			//		fmt.Println(err)
			//		return
			//	}
			//	row := make(map[string]string) //每行数据
			//	for k, v := range values { //每行数据是放在values里面，现在把它挪到row里
			//		 log.Println("value:", string(v))
			//		key := column[k]
			//		row[key] = string(v)
			//	}
			//	results[i] = row //装入结果集中
			//	i++
			//}
			//log.Println(results)

			value, ok := ct.(string)
			if ok {
				p := &admin{Username: value, Data: results}
				log.Println("r:", request.URL.Path, "当前用户：", ct)
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
func ChangeHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		err := templates.ExecuteTemplate(writer, "password.html", nil)
		log.Println("修改密码")
		if err != nil {
			log.Println(err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	} else if request.Method == "POST" {
		pass := request.FormValue("password")
		newPass := request.FormValue("reP")
		log.Println("password:", pass, "newPassword:", newPass)
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
				log.Println("修改成功", affect)
				io.WriteString(writer, "修改成功")
				//http.Redirect(writer, request, "/main/", http.StatusFound)
			}
		}

	}
}
func UserInfoHandler(writer http.ResponseWriter, request *http.Request) {

	sqline := "select  * from userinfo"
	results := make(map[int]map[string]string)
	results = query(sqline)

	//db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	//defer db.Close()
	//// fmt.Printf("SELECT password FROM userinfo where account=\"%v\"", email)
	////查询数据
	//res, err := db.Query("select  * from userinfo ")
	//
	//column, _ := res.Columns()                //读出查询出的列字段名
	//values := make([][]byte, len(column))     //values是每个列的值，这里获取到byte里
	//scans := make([]interface{}, len(column)) //因为每次查询出来的列是不定长的，用len(column)定住当次查询的长度
	//for i := range values {
	//	//让每一行数据都填充到[][]byte里面
	//	scans[i] = &values[i]
	//}
	//results := make(map[int]map[string]string) //最后得到的map
	//i := 0
	//for res.Next() {
	//	//循环，让游标往下移动
	//	if err := res.Scan(scans...); err != nil {
	//		//query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
	//		fmt.Println(err)
	//		return
	//	}
	//	row := make(map[string]string) //每行数据
	//	for k, v := range values {
	//		//每行数据是放在values里面，现在把它挪到row里
	//		// log.Println("value:", string(v))
	//		key := column[k]
	//		row[key] = string(v)
	//	}
	//	results[i] = row //装入结果集中
	//	i++
	//}
	//log.Println("r:", request.URL.Path, "results:", results)
	b, err := json.Marshal(results)
	if err == nil {
		io.WriteString(writer, string(b))
	}
	return
}
func LogoutHandler(writer http.ResponseWriter, request *http.Request) {
	globalSessions.SessionDestroy(writer, request)
	http.Redirect(writer, request, "/login/", http.StatusFound)
	return
}
func LoginHandler(writer http.ResponseWriter, request *http.Request) {
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
			} else {
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
func PostHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")

	startTime := request.FormValue("start")
	endTime := request.FormValue("end")
	log.Println("starttime:", startTime, "endtime:", endTime)

	if startTime != "" {
		id := request.FormValue("nameid")
		//uid := sess.Get("id")
		results := make(map[int]map[string]string)
		results=query("select  date,name,new_num,tol_num from userdata where date>=? and date<=? AND gameid=?",startTime, endTime, id)

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
		results := make(map[int]map[string]string)
		results=query("SELECT date,name,new_num,tol_num FROM userdata where gameid=?",id)

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
func MainHandler(writer http.ResponseWriter, request *http.Request) {
	sess := globalSessions.SessionStart(writer, request)
	ct := sess.Get("username")
	uid := sess.Get("id")
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

			db, _ := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
			defer db.Close()
			//查询数据
			rows, _ := db.Prepare("SELECT gameid , gamename  FROM gameinfo WHERE uid = ?")
			res,err:=rows.Query(uid)
			checkErr(err)
			var results map[int]string
			results = make(map[int]string)
			for res.Next() {
				var uid int
				var username string
				err := res.Scan(&uid, &username)
				checkErr(err)
				//fmt.Println(uid)
				//fmt.Println(username)
				results[uid] = username
			}

			value, ok := ct.(string)
			if ok {
				p := &admin{Username: value, Data: results}
				log.Println("r:", request.URL.Path, "当前用户：", ct)
				log.Println(results)
				er := templates.ExecuteTemplate(writer, "main.html", p)
				if er != nil {
					log.Println(er)
					http.Error(writer, er.Error(), http.StatusInternalServerError)
				}
			}
			return


			//results := make(map[int]map[string]string)
			//results = query("SELECT gameid , gamename  FROM gameinfo WHERE uid = ?",uid)
			////results = query("SELECT date,name,new_num,tol_num FROM userdata")
			//value, ok := ct.(string)
			//if ok {
			//	p := &GameData{Gamename: value, Data: results}
			//	log.Println("r:", request.URL.Path, "当前用户：", ct, "p:", p)
			//	er := templates.ExecuteTemplate(writer, "main.html", p)
			//	if er != nil {
			//		log.Println(er)
			//		http.Error(writer, er.Error(), http.StatusInternalServerError)
			//	}
			//}
			//return
		} else {
			http.Redirect(writer, request, "/login/", http.StatusFound)
			return
		}
	} else {
		http.NotFound(writer, request)
		return
	}

}
func SelectHandler(writer http.ResponseWriter, request *http.Request) {
	id := request.FormValue("uid")
	db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	defer db.Close()
	//查询数据
	res, err := db.Prepare("select gameid,gamename from gameinfo where uid=? ")
	log.Println("id:", id)
	query, err := res.Query(id)
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
	for query.Next() {
		//循环，让游标往下移动
		if err := query.Scan(scans...); err != nil {
			//query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			fmt.Println(err)
			return
		}
		row := make(map[string]string) //每行数据
		for k, v := range values {
			//每行数据是放在values里面，现在把它挪到row里
			// log.Println("value:", string(v))
			key := column[k]
			row[key] = string(v)
		}
		results[i] = row //装入结果集中
		i++
	}

	b, err := json.Marshal(results)
	if err == nil {
		//将map数组传值到HTML中
		io.WriteString(writer, string(b))
		log.Println("r:", request.URL.Path, "results:", string(b))

	}
}
func AddGameDataHandler(writer http.ResponseWriter, request *http.Request){
	gameId := request.FormValue("add_gameid")
	date := request.FormValue("date")
	new_num := request.FormValue("new_num")
	tol_num := request.FormValue("tol_num")
	fmt.Println("gameid:",gameId,"date:",date,"new_num:",new_num,"tol_num:",tol_num)
	db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	defer db.Close()
	checkErr(err)
	//查询gamename
	sql, err := db.Prepare("SELECT gamename FROM gameinfo where gameid=?")
	rows , err:=sql.Query(gameId)
	checkErr(err)
	var gamename string
	for rows.Next() {
		err = rows.Scan(&gamename)
		checkErr(err)
	}
	//插入数据
	stmt, err := db.Prepare("INSERT userdata SET gameid=?,name=?,new_num=?,tol_num=?,date=?")
	checkErr(err)

	res, err := stmt.Exec(gameId, gamename, new_num,tol_num,date)
	checkErr(err)
	if err != nil {
		fmt.Println("添加失败")
		io.WriteString(writer,"添加数据失败")
		return
	}
	fmt.Println("添加数据成功：",gamename)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)
	http.Redirect(writer, request, "/admin/", http.StatusFound)
	return
}
func AddGameHandler(writer http.ResponseWriter, request *http.Request)  {
	select_name := request.FormValue("select_name")
	new_gamename := request.FormValue("new_gamename")
	fmt.Println("new_gamename:",new_gamename,"select_name:",select_name)
	db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	defer db.Close()
	checkErr(err)
	//插入数据
	stmt, err := db.Prepare("INSERT gameinfo SET uid=?,gamename=?")
	checkErr(err)

	res, err := stmt.Exec(select_name,new_gamename)
	checkErr(err)
	if err != nil {
		fmt.Println("添加失败")
		io.WriteString(writer,"添加新游戏失败")
		return
	}
	fmt.Println("添加新游戏成功：",new_gamename)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)
	http.Redirect(writer, request, "/admin/", http.StatusFound)
	return
}
func AddUserHandler(writer http.ResponseWriter, request *http.Request)  {
	new_username := request.FormValue("new_username")
	new_account := request.FormValue("new_account")
	new_password := request.FormValue("new_password")

	fmt.Println("new_username:",new_username,"new_account:",new_account,"new_password",new_password)
	db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	defer db.Close()
	checkErr(err)
	//插入数据
	stmt, err := db.Prepare("INSERT userinfo SET username=?,account=?,password=?,admin=?")
	checkErr(err)

	res, err := stmt.Exec(new_username,new_account,new_password,0)
	checkErr(err)
	if err != nil {
		fmt.Println("添加失败")
		io.WriteString(writer,"添加新用户失败")
		return
	}
	fmt.Println("添加新游戏成功：",new_username)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)
	http.Redirect(writer, request, "/admin/", http.StatusFound)
	return
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
type GameData struct {
	Gamename string
	Data     map[int]map[string]string
}
type admin struct {
	Username string
	Data     map[int]string
}
func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}
func query(sqline ...string) map[int]map[string]string {

	fmt.Println("sqline: ", sqline)
	i := len(sqline)
	db, err := sql.Open("mysql", "root:123456@/userinfo?charset=utf8")
	checkErr(err)
	defer db.Close()
	//查询数据 "SELECT date,name,new_num,tol_num FROM userdata"
	res, err := db.Prepare(sqline[0])
	query, err := res.Query()

	if i == 2 {
		t := string(sqline[1])
		query, err = res.Query(t)
		checkErr(err)
	}else if i == 3 {
		t1 := string(sqline[1])
		t2 := string(sqline[2])
		query, err = res.Query(t1,t2)
		checkErr(err)
	}else if i == 4 {
		t1 := string(sqline[1])
		t2 := string(sqline[2])
		t3 := string(sqline[3])
		query, err = res.Query(t1,t2,t3)
		checkErr(err)
	}
	column, _ := query.Columns()              //读出查询出的列字段名
	values := make([][]byte, len(column))     //values是每个列的值，这里获取到byte里
	scans := make([]interface{}, len(column)) //因为每次查询出来的列是不定长的，用len(column)定住当次查询的长度
	for i := range values {
		//让每一行数据都填充到[][]byte里面
		scans[i] = &values[i]
	}
	results := make(map[int]map[string]string) //最后得到的map
	for query.Next() {
		//循环，让游标往下移动
		if err := query.Scan(scans...); err != nil {
			//query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			fmt.Println(err)
			return nil
		}
		row := make(map[string]string) //每行数据
		for k, v := range values {
			//每行数据是放在values里面，现在把它挪到row里
			// log.Println("value:", string(v))
			key := column[k]
			row[key] = string(v)
		}
		results[i] = row //装入结果集中
		i++
	}
	log.Println(results)
	return results
}
