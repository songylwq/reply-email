package main

import (
	"log"
	"github.com/emersion/go-imap/client"
	"time"
	"github.com/go-gomail/gomail"
	"strconv"
	"github.com/emersion/go-imap"
	"os"
	"fmt"
	"bufio"
	"io"
	"math/rand"
	"crypto/tls"
	"github.com/Unknwon/goconfig"
)

//邮件对象
type Email struct {
	Uid uint32
	FromAddr string
	ToAddr string
	Context string
}

//文本素材切片
var TxtCont = make([]string, 1)
//文本素材长度
var TxtContNum = 0
//配置文件
var Cfg *goconfig.ConfigFile

func main() {
	//初始化配置文件
	initConfig()
	//初始化文本资源
	InitTxtCont()
	findUnReadMail()
}

//初始化配置文件
func initConfig() {
	var err error
	Cfg, err = goconfig.LoadConfigFile("config/main.ini")
	if err != nil{
		log.Fatal("读取配置文件错误：", err)
	}
	inHos,_ := Cfg.GetValue("InboxMail", "host")
	log.Println("读取配置文件林成功["+inHos+"]")
}


//查询未读
func findUnReadMail() {
	log.Println("开始查询新邮件")
	inHos,_ := Cfg.GetValue("InboxMail", "host")
	inPost,_ := Cfg.GetValue("InboxMail", "port")
	c, err := client.DialTLS(inHos+":"+inPost, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("获取链接成功")
	defer c.Logout()

	inName,_ := Cfg.GetValue("InboxMail", "name")
	inPwd,_ := Cfg.GetValue("InboxMail", "pwd")

	if err := c.Login(inName, inPwd); err != nil {
		log.Fatal(err)
	}
	log.Println("登录邮箱成功")

	//每5秒查询一下收件箱，读取配置
	readInBoxTimeStr,_ := Cfg.GetValue("Sys", "readInBoxTime")
	readInBoxTime, err := strconv.ParseInt(readInBoxTimeStr, 10, 64)
	for {
		log.Println(readInBoxTimeStr+"秒后读取收件箱...")
		time.Sleep(time.Second * (time.Duration(readInBoxTime)))
		log.Println("读取收件箱")

		c.Select("INBOX", false)

		//筛选所有未读的邮件
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		uids, err := c.Search(criteria)
		if err != nil {
			log.Println(err)
		}

		if 0 < len(uids) {
			seqset := new(imap.SeqSet)
			seqset.AddNum(uids...)

			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
			messages := make(chan *imap.Message, len(uids))
			err = c.Fetch(seqset, items, messages)

			log.Println("查询收件箱完成，收件箱邮件数量：", len(uids))

			for msg := range messages {
				go func(msg *imap.Message){
					//参考报文 {"PersonalName":"\" 芝麻开花 \"","AtDomainList":"","MailboxName":"350956892","HostName":"qq.com"}
					//我的收件账户地址
					repEmailMyAddr := msg.Envelope.To[0].MailboxName + "@" + msg.Envelope.To[0].HostName
					//马甲号地址
					repEmailMajiaAddr := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
					reqEmailTitl := GetRandromTxt(15)
					reqEmailCont := GetRandromTxt(50)

					log.Println("开始回复邮件：", repEmailMajiaAddr)

					//log.Println("repEmailAddr ==>> ", repEmailAddr)
					//log.Println("reqEmailTitl ==>> ", reqEmailTitl)
					//log.Println("reqEmailCont ==>> ", reqEmailCont)

					err = SendMail(repEmailMyAddr, repEmailMajiaAddr, reqEmailTitl, reqEmailCont )
					if nil != err {
						fmt.Printf("%v",err)
						log.Fatal("邮件回复失败")
					}

					log.Println("邮件回复完成：", repEmailMajiaAddr)
				}(msg)
			}

			//将未读的邮件标记上已读
			seqset = new(imap.SeqSet)
			seqset.AddNum(uids...)
			// First mark the message as deleted
			item := imap.FormatFlagsOp(imap.AddFlags, true)
			flags := []interface{}{imap.SeenFlag}
			if err := c.Store(seqset, item, flags, nil); err != nil {
				log.Fatal(err)
			}
			log.Println("都标记成已读了")
		}

		log.Println("暂时没有未读邮件")
	}

}


//发送邮件
func SendMail(mailFrom string, mailTo string, subject string, body string) error {
	//定义邮箱服务器连接信息，如果是网易邮箱 pass填密码，qq邮箱填授权码
	//mailConn := map[string]string{
	//	"user": "songylwq@126.com",
	//	//"pass": "19861113chyylwq",
	//	"pass": "371246song",
	//	"host": "smtp.126.com",
	//	"port": "465",
	//}

	//mailConn := map[string]string{
	//	"user": "songyl",
	//	"pass": "4@2xade53",
	//	"host": "mail.wangzihan.xyz",
	//	//"host": "162.244.77.183",
	//	"port": "465",
	//}

	sendName,_ := Cfg.GetValue("SendMail", "name")
	sendPwd,_ := Cfg.GetValue("SendMail", "pwd")
	sendHost,_ := Cfg.GetValue("SendMail", "host")
	sendPort,_ := Cfg.GetValue("SendMail", "port")

	port, _ := strconv.Atoi(sendPort) //转换端口类型为int
	d := gomail.NewDialer(sendHost, port, sendName, sendPwd)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	sendCloser,err := d.Dial()
	if err != nil {
		fmt.Printf("创建sendCloser异常：%s\n", err.Error())
	}
	log.Println(mailTo+"-创建sendCloser成功")

	for i:=0; i<1; i++ {
		m := gomail.NewMessage()
		m.SetAddressHeader("From", mailFrom, mailFrom)
		m.SetAddressHeader("To", mailTo, mailTo)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", "<h1>"+body+"</h1>"+"<span>"+body+body+"</span>")
		sendCloser.Send(mailFrom, []string{mailTo}, m)
		log.Println("发送邮件from["+mailFrom+"],to["+mailTo+"]成功")
	}

	return err
}

//初始化文本内容，读取到内存里面
func InitTxtCont(){
	filePathUrl := fmt.Sprintf("data%ctxtcontent", os.PathSeparator)
	log.Printf("开始读取[%s]", filePathUrl)
	ReadTxtCont(filePathUrl, 10000, TxtHandle)
	TxtContNum = len(TxtCont)
	log.Printf("读取文本素材完成,共读取[%v]条", len(TxtCont))
}

//将文本内容读取到内存里
func TxtHandle(line []byte) {
	//os.Stdout.Write(line)
	//fmt.Println(string(line))
	//将文本素材写入内存
	TxtCont = append(TxtCont, string(line))
}

//读取文本内容
func ReadTxtCont(filePth string, bufSize int, hookfn func([]byte)) error {
	f, err := os.Open(filePth)
	if err != nil {
		return err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f)
	for {
		line, err := bfRd.ReadBytes('\n')
		hookfn(line) //放在错误处理前面，即使发生错误，也会处理已经读取到的数据。
		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

//随机获取一个指定长度的文本
func GetRandromTxt(getCount int) string {
	//获取随机数
	randZh := rand.New(rand.NewSource(time.Now().UnixNano()))
	randId := randZh.Intn(TxtContNum)
	reStr := TxtCont[randId]
	for getCount > len([]rune(reStr)) {
		randZh = rand.New(rand.NewSource(time.Now().UnixNano()))
		randId = randZh.Intn(TxtContNum)
		reStr += TxtCont[randId]
	}

	nameRune := []rune(reStr)

	return string(nameRune[1 : getCount])
}


//func main() {
//	log.Println("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	// List mailboxes
//	mailboxes := make(chan *imap.MailboxInfo, 10)
//	done := make(chan error, 1)
//	go func () {
//		done <- c.List("", "*", mailboxes)
//	}()
//
//	log.Println("Mailboxes:")
//	for m := range mailboxes {
//		log.Println("* " + m.Name)
//	}
//
//	log.Println("******")
//
//	if err := <-done; err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("Select INBOX")
//
//	// Select INBOX
//	mbox, err := c.Select("INBOX", false)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Flags for INBOX:", mbox.Flags)
//
//	// Get the last 4 messages
//	from := uint32(1)
//	to := mbox.Messages
//	if mbox.Messages > 5 {
//		// We're using unsigned integers here, only substract if the result is > 0
//		from = mbox.Messages - 5
//	}
//	seqset := new(imap.SeqSet)
//	seqset.AddRange(from, to)
//
//	log.Println("mbox.Messages:", mbox.Messages)
//
//	messages := make(chan *imap.Message, 10)
//	done = make(chan error, 1)
//	go func() {
//		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
//	}()
//
//	log.Println("Last 4 messages:")
//	for msg := range messages {
//		log.Println("* " + msg.Envelope.Subject)
//		//log.Println("from: ", msg.Envelope.From[0])
//		//log.Println("to: ", msg.Envelope.To[0])
//		//log.Println("ReplyTo: ", msg.Envelope.ReplyTo[0])
//	}
//
//	if err := <-done; err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("Done!")
//}


//func main() {
//	log.Println("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	//// List mailboxes
//	//mailboxes := make(chan *imap.MailboxInfo, 10)
//	//done := make(chan error, 1)
//	//go func () {
//	//	done <- c.List("", "*", mailboxes)
//	//	//done <- c.List("", "INBOX", mailboxes)
//	//}()
//	//
//	//log.Println("Mailboxes:")
//	//for m := range mailboxes {
//	//	log.Println("* " + m.Name)
//	//}
//	//
//	//log.Println("******")
//	//
//	//if err := <-done; err != nil {
//	//	log.Fatal(err)
//	//}
//	//
//	//log.Println("Select INBOX")
//	//
//	//// Select INBOX
//	//mbox, err := c.Select("INBOX", false)
//	//
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//
//	//log.Println("Flags for INBOX:", mbox.Flags)
//	//
//	////mbox.Flags = []string{"\\Deleted"}
//	////mbox.PermanentFlags = []string{"\\Deleted"}
//	////
//	////log.Println("Flags 修改后:", mbox.Flags)
//	////log.Println("PermanentFlags 修改后:", mbox.PermanentFlags)
//	//
//	//// Get the last 4 messages
//	//from := uint32(1)
//	//to := mbox.Messages
//	//if mbox.Messages > 2 {
//	//	// We're using unsigned integers here, only substract if the result is > 0
//	//	from = mbox.Messages - 2
//	//}
//	//seqset := new(imap.SeqSet)
//	//seqset.AddRange(from, to)
//	//
//	//log.Println("mbox.Messages : ", mbox.Messages)
//	//log.Println("mbox.PermanentFlags : ", mbox.PermanentFlags)
//
//
//
//
//	//done := make(chan error, 1)
//
//	for i:=1;i<10;i++{
//		//读取服务器的收件箱邮件的数量，确定读取邮件的起止角标
//		mbox, err := c.Select("INBOX", false)
//
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		from := uint32(1)
//		to := mbox.Messages
//		if mbox.Messages > 2 {
//			// We're using unsigned integers here, only substract if the result is > 0
//			from = mbox.Messages - 2
//		}
//		seqset := new(imap.SeqSet)
//		seqset.AddRange(from, to)
//
//		messages := make(chan *imap.Message, 10)
//		err = c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}, messages)
//
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		time.Sleep(3 * time.Second)
//
//		log.Println()
//		log.Println("Last 4 messages:")
//		for msg := range messages {
//			log.Println("* " + msg.Envelope.Subject)
//			log.Println("*,Flags : ", msg.Flags)
//			//log.Println("from: ", msg.Envelope.From[0])
//			//log.Println("to: ", msg.Envelope.To[0])
//			//log.Println("ReplyTo: ", msg.Envelope.ReplyTo[0])
//			//log.Println("MessageId: ", msg.Envelope.MessageId)
//		}
//
//	}
//
//	//回复邮件
//
//
//	log.Println("Done!")
//}


//
////删除邮件
//func main() {
//	log.Println("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	// Select INBOX
//	mbox, err := c.Select("INBOX", false)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// We will delete the last message
//	if mbox.Messages == 0 {
//		log.Fatal("No message in mailbox")
//	}
//
//	seqset := new(imap.SeqSet)
//	seqset.AddNum(mbox.Messages)
//
//	// First mark the message as deleted
//	item := imap.FormatFlagsOp(imap.AddFlags, true)
//	flags := []interface{}{imap.SeenFlag}
//	if err := c.Store(seqset, item, flags, nil); err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("删除了")
//}


////查询未读邮件
//func main() {
//	log.Println("Connecting to server...")
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//	// Don't forget to logout
//	defer c.Logout()
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	for {
//		c.Select("test", false)
//
//		//筛选所有未读的邮件
//		criteria := imap.NewSearchCriteria()
//		criteria.WithoutFlags = []string{"\\Seen"}
//		uids, err := c.Search(criteria)
//		if err != nil {
//			log.Println(err)
//		}
//
//		log.Println(uids)
//
//		if 0 < len(uids) {
//			seqset := new(imap.SeqSet)
//			seqset.AddNum(uids...)
//
//			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
//			messages := make(chan *imap.Message, len(uids))
//			err = c.Fetch(seqset, items, messages)
//
//			log.Println("c.Fetched")
//
//			for msg := range messages {
//				log.Println("* ", msg.Envelope.Subject, "--",  msg.Flags)
//			}
//
//			//将未读的邮件标记上已读
//			seqset = new(imap.SeqSet)
//			seqset.AddNum(uids...)
//			// First mark the message as deleted
//			item := imap.FormatFlagsOp(imap.AddFlags, true)
//			flags := []interface{}{imap.SeenFlag}
//			if err := c.Store(seqset, item, flags, nil); err != nil {
//				log.Fatal(err)
//			}
//			log.Println("都标记成已读了")
//		}
//
//		log.Println("暂时没有未读邮件")
//		time.Sleep(time.Second * 5)
//	}
//}


