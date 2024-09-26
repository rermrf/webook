package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	// Author 要从用户来
	Author Author
	Status ArticleStatus
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
