package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	// Author 要从用户来
	Author Author
}

// Author 在帖子这个领域内是一个值对象
type Author struct {
	Id   int64
	Name string
}
