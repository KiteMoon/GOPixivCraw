package main

type ListDataErr struct {
	Error string `json:"error"`
}

// 解析Pixiv用结构体
type ListData struct {
	Mode      string         `json:"mode"`
	Content   string         `json:"content"`
	Contents  []ContentsData `json:"contents"`
	Page      int            `json:"page"`
	Prev      interface{}    `json:"prev"` //这个接口有点鬼畜，你请求第一页会返回false，你请求第二页会返回1？？？？
	Next      int            `json:"next"`
	Date      string         `json:"date"`
	PrevDate  string         `json:"prev_date"`
	NextDate  bool           `json:"next_date"`
	RankTotal int            `json:"rank_total"`
}

// 解析Pixiv用子结构体
type ContentsData struct {
	Title                 string      `json:"title"`
	Date                  string      `json:"date"`
	Tags                  []string    `json:"tags"`
	Url                   string      `json:"url"`
	IllustType            string      `json:"illust_type"`
	IllustBookStyle       string      `json:"illust_book_style"`
	IllustPageCount       string      `json:"illust_page_count"`
	UserName              string      `json:"user_name"`
	ProfileImg            string      `json:"profile_img"`
	IllustContentType     interface{} `json:"illust_content_type"`
	IllustSeries          interface{} `json:"illust_series"`
	IllustId              int         `json:"illust_id"`
	Width                 int         `json:"width"`
	Height                int         `json:"height"`
	UserId                int         `json:"user_id"`
	Rank                  int         `json:"rank"`
	YesRank               int         `json:"yes_rank"`
	RatingCount           int         `json:"rating_count"`
	ViewCount             int         `json:"view_count"`
	IllustUploadTimestamp int         `json:"illust_upload_timestamp"`
	Attr                  string      `json:"attr"`
}

// 推送模块结构体
type PushData struct {
	Token   string `json:"token"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// 数据库入库结构体
