package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	pushPlusToken string = os.Getenv("pushplustoken")
)

func main() {
	fmt.Println("helloworld")
	page := 3
	code, meessage := GetPixivList(page)
	PushErrorMessage(code, meessage)
}

// 实现一个全局请求器
func GetPixivList(page int) (code, error string) {
	//构建请求
	//写个循环，循环要榜单前多少
	//var data string
	for i := 1; i <= page; i++ {

		pageUrl := fmt.Sprintf("https://www.pixiv.net/ranking.php?p=%d&format=json", i)
		pageUrl = "https://www.pixiv.net/ranking.php?p=x&format=json"
		fmt.Println(pageUrl)
		pixivListRequestsClient := http.Client{}
		pixivListRequests, err := http.NewRequest("GET", pageUrl, nil)
		if err != nil {
			return "400", "构建请求失败，请检查请求参数\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_001"
		}
		pixivResponse, err := pixivListRequestsClient.Do(pixivListRequests)
		if err != nil {
			return "400", "发起请求失败，请检查网络参数\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_002"

		}
		// 先判断下是不是大型错误（比如直接返回一个html）
		pixivResponseDataString, err := ioutil.ReadAll(pixivResponse.Body)
		if err != nil {

			return "403", "处理请求失败，Pixiv返回了无法被识别的信息(怀疑是参数问题)\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_003"
		}
		//fmt.Println(string(pixivResponseDataString))
		checkData := ListDataErr{}
		err = json.Unmarshal(pixivResponseDataString, &checkData)
		if err != nil {
			return "403", "处理请求失败，读取Pixiv预检测失败\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_004"
		}

		if checkData.Error != "" {

			return "403", "处理请求失败，Pixiv预检测失败,发现错误参数\n错误信息为：" + checkData.Error + "\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_004"

		}
		data := ListData{}
		err = json.Unmarshal(pixivResponseDataString, &data)
		if err != nil {
			fmt.Println(err)
			return "403", "处理请求失败，无法处理通过预检后的数据，可能是Pixiv修改了API\n暂停本轮请求\n错误代码:ERROR_GET_PIXIV_004"
		}
		fmt.Printf("当前页码为%d\n:", i)
		fmt.Println("本页采总共采集到以下图片")
		for t := 0; t < len(data.Contents); t++ {
			fmt.Printf("图片序号:%d\t", t+1)
			fmt.Println(data.Contents[t].Title)
		}
	}

	return "200", "没有发生错误"
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
