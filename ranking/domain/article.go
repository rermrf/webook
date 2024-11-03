package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	// Author 要从用户来
	Author Author
	Status ArticleStatus
	Ctime  time.Time
	Utime  time.Time

	// 做成这样，就应该在 service 或者 repository 里面完成构造
	// 设计成这个样子，就认为 Interactive 是 Article 的一个属性（值对象）
	// Intr Interactive
}

func (art Article) Abstract() string {
	// 摘要取前几句
	// 要考虑中文问题
	cs := []rune(art.Content)
	if len(cs) < 100 {
		return art.Content
	}
	// 英文怎么截取一个完整的单词，不需要纠结，就截断拉到
	// 词组、介词，往后找标点符号
	return string(cs[:100]) + "..."
}

type ArticleStatus uint8

const (
	// ArticleStatusUnknown 为了避免零值之类的问题
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnPublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) Valid() bool {
	return s.ToUint8() > 0
}

func (s ArticleStatus) NonPublish() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusUnPublished:
		return "unPublished"
	case ArticleStatusPublished:
		return "published"
	case ArticleStatusPrivate:
		return "private"
	default:
		return "unknown"
	}
}

// 如果你的状态很复杂，有很多行为（就是要搞很多方法）
type ArticleStatusV1 struct {
	Val  uint8
	Name string
}

var (
	ArticleStatusV1Unknown = ArticleStatusV1{Val: 0, Name: "unknown"}
)

// Author 在帖子这个领域内是一个值对象
type Author struct {
	Id   int64
	Name string
}
