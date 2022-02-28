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
	IllustId              int64       `json:"illust_id"`
	Width                 int64       `json:"width"`
	Height                int64       `json:"height"`
	UserId                int64       `json:"user_id"`
	Rank                  int64       `json:"rank"`
	YesRank               int64       `json:"yes_rank"`
	RatingCount           int64       `json:"rating_count"`
	ViewCount             int64       `json:"view_count"`
	IllustUploadTimestamp int64       `json:"illust_upload_timestamp"`
	Attr                  string      `json:"attr"`
}

// 推送模块结构体
type PushData struct {
	Token   string `json:"token"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// 数据库入库结构体
type SqlData struct {
	PID         int64
	TOPNUM      int64
	TITLE       string
	ANTHOR      string
	ANTHORID    int64
	TOPDATE     int64
	UPLOADDATE  string
	TAG         string
	PHOTONUM    string
	PHOTOWIDTH  int64
	PHOTOHEIGHT int64
	VIEWURL     string
	TOPTREND    string
}
