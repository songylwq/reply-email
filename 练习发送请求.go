package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"time"
	"encoding/json"
	"net/url"
	"strings"
	"strconv"
)

func main() {
	//var mutex sync.Mutex
	for i:=1; i<10; i++ {

		go func(i int){
			//mutex.Lock()
			//fmt.Println("获取到i:",i)
			//httpGet(strconv.Itoa(i))
			//mutex.Unlock()


			var r http.Request
			r.ParseForm()
			r.Form.Add("username", "admin")
			r.Form.Add("password", "123456")
			bodystr := strings.TrimSpace(r.Form.Encode())
			request, err := http.NewRequest("GET", "http://localhost:12101/login2?requId="+strconv.Itoa(i),
				strings.NewReader(bodystr))

			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Connection", "Keep-Alive")


			client := &http.Client{}

			client.Timeout = time.Second * 5

			resp, err := client.Do(request)

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				// handle error
			}

			fmt.Println(string(body))


		}(i)

	}
	time.Sleep(100*time.Second)
}

func httpGet(id string) {
	resp, err := http.PostForm("http://localhost:12101/login2?requId="+id,
		url.Values{"username": {"admin"}, "password": {"123456"}})

	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println("body:", string(body))

	respNy := make(map[string]string)
	json.Unmarshal(body, &respNy)

	if "200" == respNy["Code"] {
		fmt.Println("ok")
	}
}