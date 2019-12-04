package main

import (
	"crypto/tls"
	"fmt"

	"github.com/go-gomail/gomail"
)

func main() {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "songyl@wangzihan.xyz", "账号1")
	m.SetAddressHeader("To", "songylwq@126.com", "账号2")
	m.SetHeader("Subject", "gomail-邮件测试")
	m.SetBody("text/html", "<h1>hello world</h1>")

	d := gomail.NewDialer("mail.wangzihan.xyz", 465, "songyl", "4@2xade53")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("***%s\n", err.Error())
	}
}