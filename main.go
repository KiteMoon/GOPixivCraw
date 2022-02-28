package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	pushPlusToken string = os.Getenv("pushplustoken")
	err           error
	DB            *sql.DB
)

// 初始化
func init() {
	dsn := ""
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("发生全局错误，数据库连接失败")
		panic(err)
		return
	}
	err = DB.Ping()
	if err != nil {
		fmt.Println("与数据库连接终端，心跳错误，抛出错误")
		panic(err)
	}
	fmt.Println("hihi")
}
func main() {
	fmt.Println("helloworld")
	page := 2
	var bigdata []SqlData
	for i := 0; i < page; i++ {
		code, message, data := GetPixivList(i + 1)
		PushErrorMessage(code, message)
		bigdata = append(bigdata, data...)
	}
	fmt.Println(bigdata)
	for t := 0; t < len(bigdata); t++ {
		test := "insert into top_list(PID,TOPNUM,TITLE,ANTHOR,ANTHORID,TOPDATE,UPLOADDATE,TAG,PHOTONUM,PHOTOWIDTH,PHOTOHEIGHT,VIEWURL,TOPTREND) values (?,?,?,?,?,?,?,?,?,?,?,?,?)"
		_, err = DB.Exec(test, bigdata[t].PID, bigdata[t].TOPNUM, bigdata[t].TITLE, bigdata[t].ANTHOR, bigdata[t].ANTHORID, bigdata[t].TOPDATE, bigdata[t].UPLOADDATE, bigdata[t].TAG, bigdata[t].PHOTONUM, bigdata[t].PHOTOWIDTH, bigdata[t].PHOTOHEIGHT, bigdata[t].VIEWURL, "开发字段，暂不开放")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}

// 实现一个全局请求器
func GetPixivList(page int) (code, error string, data []SqlData) {
	//构建请求
	//写个循环，循环要榜单前多少
	//var data string
	i := page
	pageUrl := fmt.Sprintf("https://www.pixiv.net/ranking.php?p=%d&format=json", i)
	//pageUrl = "https://www.pixiv.net/ranking.php?p=x&format=json"
	fmt.Println(pageUrl)
	pixivListRequestsClient := http.Client{}
	pixivListRequests, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return "400", "构建请求失败，请检查请求参数\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_001", []SqlData{}
	}
	pixivResponse, err := pixivListRequestsClient.Do(pixivListRequests)
	if err != nil {
		return "400", "发起请求失败，请检查网络参数\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_002", []SqlData{}

	}
	// 先判断下是不是大型错误（比如直接返回一个html）
	pixivResponseDataString, err := ioutil.ReadAll(pixivResponse.Body)
	if err != nil {

		return "403", "处理请求失败，Pixiv返回了无法被识别的信息(怀疑是参数问题)\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_003", []SqlData{}
	}
	//fmt.Println(string(pixivResponseDataString))
	checkData := ListDataErr{}
	err = json.Unmarshal(pixivResponseDataString, &checkData)
	if err != nil {
		return "403", "处理请求失败，读取Pixiv预检测失败\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_004", []SqlData{}
	}

	if checkData.Error != "" {

		return "403", "处理请求失败，Pixiv预检测失败,发现错误参数\n错误信息为：" + checkData.Error + "\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_004", []SqlData{}

	}
	respondata := ListData{}
	err = json.Unmarshal(pixivResponseDataString, &respondata)
	if err != nil {
		fmt.Println(err)
		return "403", "处理请求失败，无法处理通过预检后的数据，可能是Pixiv修改了API\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_004", []SqlData{}
	}
	fmt.Printf("当前页码为%d\n:", i)
	fmt.Println("本页采总共采集到以下图片")
	var sqldata []SqlData
	for t := 0; t < len(respondata.Contents); t++ {
		var sqldata1 SqlData

		fmt.Printf("图片序号:%d\t", t+1)
		fmt.Println(respondata.Contents[t].Title)
		fmt.Println(respondata.Contents[t].IllustId)
		sqldata1.PID = respondata.Contents[t].IllustId
		sqldata1.TOPNUM = respondata.Contents[t].Rank
		sqldata1.TITLE = respondata.Contents[t].Title
		sqldata1.ANTHOR = respondata.Contents[t].UserName
		sqldata1.ANTHORID = respondata.Contents[t].UserId
		sqldata1.TOPDATE = respondata.Contents[t].IllustUploadTimestamp
		sqldata1.UPLOADDATE = respondata.Contents[t].Date
		for x := 0; x < len(respondata.Contents[t].Tags); x++ {
			if x == len(respondata.Contents[t].Tags)-1 {
				sqldata1.TAG = sqldata1.TAG + respondata.Contents[t].Tags[x]
				continue
			}
			sqldata1.TAG = sqldata1.TAG + respondata.Contents[t].Tags[x] + ","

		}
		sqldata1.PHOTONUM = respondata.Contents[t].IllustPageCount
		sqldata1.PHOTOWIDTH = respondata.Contents[t].Width
		sqldata1.PHOTOHEIGHT = respondata.Contents[t].Height
		sqldata1.VIEWURL = respondata.Contents[t].Url
		sqldata1.TOPTREND = "test"
		sqldata = append(sqldata, sqldata1)
	}

	return "200", "没有发生错误", sqldata
}

// 实现一个error处理函数
func PushErrorMessage(code string, message string) {
	//只提供pushPlus接口，请
	pushUrl := "http://www.pushplus.plus/send"
	pushClient := http.Client{}
	if code != "200" {
		title := fmt.Sprintf("发生错误，错误代码:%s", code)
		message := fmt.Sprintf("发生错误，错误信息如下\n%s", message)
		requestData := PushData{
			Token:   pushPlusToken,
			Title:   title,
			Content: message,
		}
		requestDataJson, _ := json.Marshal(requestData)

		pushRequest, _ := http.NewRequest("POST", pushUrl, bytes.NewBuffer(requestDataJson))
		response, err := pushClient.Do(pushRequest)
		fmt.Println(response)
		if err != nil {
			fmt.Println("发送push失败")
			return
		}

	}

}

func WriteDataBase() {

}

// 实现一个消息队列
func AddPixiv() {

}
