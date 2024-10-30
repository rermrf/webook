package repository

import "context"

type HistoryRecordRepository interface {
	AddRecord(background context.Context, biz string, bizId int64) error
}
