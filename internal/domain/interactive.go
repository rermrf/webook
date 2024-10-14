package domain

type Interactive struct {
	//Biz        string
	//BizId      int64
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Liked      bool
	Collected  bool
}

//type Collection struct {
//	Name  string
//	Uid   int64
//	Items []Resource
//}

// max(发送者总速率/单一分区写入速率，发送者总速率/单一消费速率) + buffer
