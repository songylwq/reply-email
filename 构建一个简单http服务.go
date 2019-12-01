package main

import (
	"net/http"
	"encoding/json"
	"log"
	"fmt"
)

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/login2", login2)

	log.Println("Starting v2 httpserver")
	log.Fatal(http.ListenAndServe(":12101", mux))
	fmt.Println("开始")
}



type Resp struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
}

var requestCount = 0
//var mutex sync.Mutex

//接收x-www-form-urlencoded类型的post请求或者普通get请求
func login2(writer http.ResponseWriter,  request *http.Request)  {
	fmt.Println("获取到请求:",request.URL)
	//mutex.Lock()
	requestCount++
	//mutex.Unlock()
	request.ParseForm()
	username, uError :=  request.Form["username"]
	pwd, pError :=  request.Form["password"]

	// 获取请求报文的内容长度
	len := request.ContentLength

	// 新建一个字节切片，长度与请求报文的内容长度相同
	body := make([]byte, len)

	// 读取 r 的请求主体，并将具体内容读入 body 中
	request.Body.Read(body)
	fmt.Println(string(body))

	var result = new(Resp)
	if !uError || !pError {
		result.Code = "401"
		result.Msg = "登录失败"
	} else if username[0] == "admin" && pwd[0] == "123456" {
		result.Code = "200"
		result.Msg = "登录成功"
	} else {
		result.Code = "203"
		result.Msg = "账户名或密码错误"
	}
	if err := json.NewEncoder(writer).Encode(result); err != nil {
		log.Fatal(err)
	}
	fmt.Println("requestCount:", requestCount)
}