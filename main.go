package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	pushPlusToken string
	err           error
	pushTime      int64
	DB            *sql.DB
)

// 初始化
func init() {
	viper.SetConfigFile("./config/config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("致命错误")
		fmt.Println("本次错误无法触发推送")
		panic(fmt.Errorf("读取配置文件失败: %s \n", err))
	}
	host := viper.GetString("config.mysql.host")
	dbname := viper.GetString("config.mysql.dbname")
	username := viper.GetString("config.mysql.username")
	password := viper.GetString("config.mysql.password")
	pushTime = viper.GetInt64("config.startTS")
	pushType := viper.GetString("config.pushconfig.type")
	if pushType == "push++" {
		pushPlusToken = viper.GetString("config.pushconfig.token")
	} else {
		fmt.Println("未知的推送渠道，错误")
		panic("NO PUSH")
	}
	fmt.Println("数据库地址:", host)
	fmt.Println("数据库名称:", dbname)
	fmt.Println("数据库的用户名:", username)
	fmt.Println("数据库的密码", password)
	fmt.Println("推送起始时间戳", pushTime)
	fmt.Println("推送方式", pushType)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, dbname)
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
	DB.SetConnMaxLifetime(30)
}
func main() {
	fmt.Println("开始启动爬虫")
	page := 10
	var bigdata []SqlData
	for i := 0; i < page; i++ {
		code, message, data := GetPixivList(i + 1)
		PushErrorMessage(code, message)
		bigdata = append(bigdata, data...)
	}
	defer DB.Close()
	for t := 0; t < len(bigdata); t++ {
		code := QueryPidList(bigdata[t].PID, bigdata[t].TOPNUM)
		if code == 200 {
			fmt.Printf("\n---监控程序---\nPID:%d\n该PID对应的作品已经存在于数据库 \n------\n\n", bigdata[t].PID)
			continue
		}
		test := "insert into top_list(PID,FIRSTTOPNUM,TOPNUM,MOSTTOPNUM,MODTIME,TITLE,ANTHOR,ANTHORID,TOPDATE,UPLOADDATE,TAG,PHOTONUM,PHOTOWIDTH,PHOTOHEIGHT,VIEWURL,TOPTREND) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		_, err = DB.Exec(test, bigdata[t].PID, bigdata[t].TOPNUM, bigdata[t].TOPNUM, bigdata[t].TOPNUM, time.Now().Unix(), bigdata[t].TITLE, bigdata[t].ANTHOR, bigdata[t].ANTHORID, bigdata[t].TOPDATE, bigdata[t].UPLOADDATE, bigdata[t].TAG, bigdata[t].PHOTONUM, bigdata[t].PHOTOWIDTH, bigdata[t].PHOTOHEIGHT, bigdata[t].VIEWURL, bigdata[t].TOPTREND)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}

// 实现一个全局请求器
func GetPixivList(page int) (code, error string, data []SqlData) {
	//请勿加拿大自爆兵式请求
	//time.Sleep(10 * time.Second)
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
		sqldata1.TOPTREND = fmt.Sprintf("[{\"RANK\":%d,\"TIME\":%d}]", respondata.Contents[t].Rank, time.Now().Unix())
		sqldata = append(sqldata, sqldata1)
	}

	return "200", "没有发生错误", sqldata
}

// 实现一个error处理函数
func PushErrorMessage(code string, message string) {
	nowTime := time.Now().Unix()
	if nowTime-pushTime < 600 {
		fmt.Println("十分钟内推送过一次，自动停止推送")
		return
	}
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
		_, err := pushClient.Do(pushRequest)
		if err != nil {
			fmt.Println("发送push失败")
			return
		}
		fmt.Println("发起推送成功")
		pushTime = time.Now().Unix()

	}

}

func WriteDataBase() {

}

// 实现一个全局检查器，检查是否被收录过
func QueryPidList(pid, topnum int64) (code int) {
	fmt.Println("---数据库程序---")
	fmt.Printf("PID:%d\n", pid)
	defer fmt.Printf("\n------")
	querySql := "SELECT TITLE,TOPTREND,TOPNUM FROM TOP_LIST WHERE PID = ?"
	var data SqlDataQuery
	err := DB.QueryRow(querySql, pid).Scan(&data.TITLE, &data.TOPTREND, &data.TOPNUM)
	if err != nil {
		fmt.Printf("该作品为第一次上榜")
		return 404
	}
	if data.TOPNUM == topnum {
		fmt.Printf("该作品已经进入数据库，但是排名未变化")
		return 200
	}
	var topRankTrend []PixivPidRankTrend
	toptrend := data.TOPTREND

	err = json.Unmarshal([]byte(data.TOPTREND), &topRankTrend)
	if err != nil {
		fmt.Println("解析趋势失败，跳过趋势解析")

	} else {
		// 这里提供趋势追加
		fmt.Println(topnum)
		r := PixivPidRankTrend{
			RANK: topnum,
			TIME: time.Now().Unix(),
		}
		topRankTrend = append(topRankTrend, r)
		newTopRankTrend, err := json.Marshal(topRankTrend)
		if err != nil {
			fmt.Println("解析趋势失败，跳过趋势解析")

		} else {
			toptrend = string(newTopRankTrend)
		}
	}
	fmt.Println("该作品排名发生变化")
	fmt.Println("数据库记录排名:", data.TOPNUM)
	fmt.Println("在线排名:", topnum)
	fmt.Println("信息正在修改")
	if topnum < data.TOPNUM {
		fmt.Println("排名上升")
		updateSql := "update TOP_LIST SET TOPTREND = ? ,TOPNUM = ? ,MOSTTOPNUM=? ,MODTIME=? WHERE PID = ?"
		_, err = DB.Exec(updateSql, toptrend, topnum, topnum, time.Now().Unix(), pid)
		if err != nil {
			fmt.Println("修改数据库失败，已经发起报警")
			fmt.Println(err)
			PushErrorMessage("400", err.Error())
			fmt.Println("报警成功")
			return
		} else {
			fmt.Println("数据库修改成功，完成记录")
			return 200
		}
	}
	updateSql := "update TOP_LIST SET TOPTREND = ?   ,TOPNUM = ? ,MODTIME=? WHERE PID = ?"
	_, err = DB.Exec(updateSql, toptrend, topnum, time.Now().Unix(), pid)
	if err != nil {
		fmt.Println("修改数据库失败，已经发起报警")
		fmt.Println(err)
		PushErrorMessage("400", err.Error())
		fmt.Println("报警成功")
		return
	} else {
		fmt.Println("数据库修改成功，完成记录")
		return 200
	}
}
