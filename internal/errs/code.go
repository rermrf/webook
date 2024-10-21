package errs

// 用户模块
const (
	// UserInvalidInput 用户模块输入错误，就是一个含糊的错误
	UserInvalidInput = 401001
	// UserInvalidOrPassword 用户不存在或者密码错误
	UserInvalidOrPassword   = 401002
	UserInternalServerError = 501001
)

const (
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)

//var (
//	UserInvalidInputV1 = Code{
//		Number: 401001,
//		Msg:    "用户输入错误",
//	}
//)
//
//type Code struct {
//	Number int
//	Msg    string
//}
