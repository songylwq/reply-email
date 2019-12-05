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

func main() {
	//初始化文本资源
	InitTxtCont()
	findUnReadMail()

	//err := SendMail("songylwq@126.com",
	//	"你好的否的",
	//	"你好的否的")
	//if nil != err {
	//	log.Fatal("邮件回复失败：", err)
	//}
	//
	//log.Printf("邮件发送完成[%v]")
}

//查询未读
func findUnReadMail() {
	log.Println("开始查询新邮件")
	c, err := client.DialTLS("mail.wangzihan.xyz:993", nil)
	//c, err := client.Dial("mail.wangzihan.xyz:143")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("获取链接成功")
	defer c.Logout()
	// Login
	//if err := c.Login("songylwq@126.com", "371246song"); err != nil {
	//	log.Fatal(err)
	//}
	if err := c.Login("songyl", "4@2xade53"); err != nil {
		log.Fatal(err)
	}
	log.Println("登录邮箱成功")

	//每5秒查询一下收件箱
	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		log.Println("读取收件箱")

		c.Select("INBOX", false)

		//筛选所有未读的邮件
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		uids, err := c.Search(criteria)
		if err != nil {
			log.Println(err)
		}

		log.Println(uids)

		if 0 < len(uids) {
			seqset := new(imap.SeqSet)
			seqset.AddNum(uids...)

			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
			messages := make(chan *imap.Message, len(uids))
			err = c.Fetch(seqset, items, messages)

			log.Println("c.Fetched")

			for msg := range messages {
				log.Println("* ", msg.Envelope.Subject, "--",  msg.Flags)
				log.Println("===>> ", msg.Envelope.From[0])
				//参考报文 {"PersonalName":"\" 芝麻开花 \"","AtDomainList":"","MailboxName":"350956892","HostName":"qq.com"}
				repEmailAddr := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
				reqEmailTitl := GetRandromTxt(15)
				reqEmailCont := "你好"+msg.Envelope.From[0].PersonalName+",邮件已经收到！！\n"+GetRandromTxt(30)

				log.Println("repEmailAddr ==>> ", repEmailAddr)
				log.Println("reqEmailTitl ==>> ", reqEmailTitl)
				log.Println("reqEmailCont ==>> ", reqEmailCont)

				err = SendMail(repEmailAddr, reqEmailTitl, reqEmailCont )
				if nil != err {
					fmt.Printf("%v",err)
					log.Fatal("邮件回复失败")
				}

				log.Printf("邮件回复完成[%v]",repEmailAddr)
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
func SendMail(mailTo string, subject string, body string) error {
	//定义邮箱服务器连接信息，如果是网易邮箱 pass填密码，qq邮箱填授权码
	//mailConn := map[string]string{
	//	"user": "songylwq@126.com",
	//	//"pass": "19861113chyylwq",
	//	"pass": "371246song",
	//	"host": "smtp.126.com",
	//	"port": "465",
	//}
	mailConn := map[string]string{
		"user": "songyl",
		"pass": "4@2xade53",
		"host": "mail.wangzihan.xyz",
		//"host": "162.244.77.183",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()
	m.SetAddressHeader("From", "songyl@wangzihan.xyz", "账号1")
	m.SetAddressHeader("To", mailTo, "账号2")
	m.SetHeader("Subject", "gomail-"+subject)
	m.SetBody("text/html", "<h1>"+body+"</h1>")

	d := gomail.NewDialer("mail.wangzihan.xyz", port, "songyl", "4@2xade53")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := d.DialAndSend(m)
	//if err != nil {
	//	fmt.Printf("***%s\n", err.Error())
	//}
	//fmt.Printf(mailTo+"-发送邮件成功\n")

	sendCloser,err := d.Dial()
	if err != nil {
		fmt.Printf("***%s\n", err.Error())
	}
	fmt.Printf(mailTo+"-创建sendCloser成功\n")

	for i:=0; i<1; i++ {
		sendCloser.Send("songyl@wangzihan.xyz", []string{mailTo}, m)
		fmt.Printf(mailTo+"-发送邮件[%v]成功\n", i)
		time.Sleep(time.Second * 2 )
	}

	//m := gomail.NewMessage()
	//
	//m.SetHeader("From", m.FormatAddress(mailConn["user"], "XX官方")) //这种方式可以添加别名，即“XX官方”
	////说明：如果是用网易邮箱账号发送，以下方法别名可以是中文，如果是qq企业邮箱，以下方法用中文别名，会报错，需要用上面此方法转码
	////m.SetHeader("From", "FB Sample"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“FB Sample”，
	//// 也可以直接用m.SetHeader("From",mailConn["user"]) 读者可以自行实验下效果
	////m.SetHeader("From", mailConn["user"])
	//m.SetHeader("To", mailTo) //发送给多个用户
	//m.SetHeader("Subject", subject) //设置邮件主题
	//m.SetBody("text/html", body) //设置邮件正文
	//
	//log.Println("11111")
	//d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	//log.Println("22222")
	//err := d.DialAndSend(m)
	//log.Println("3333")
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
	for getCount > len(reStr) {
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


