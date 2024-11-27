package events

type SyncDataEvent struct {
	IndexName string
	DocId     string
	// 这里应该是 BizTags
	Data string
}

//type SyncDataEventConsumer struct {
//	svc service.Sy
//}
