package domain

// Template 通知模板
type Template struct {
	Id                    int64
	TemplateId            string
	Channel               Channel
	Name                  string
	Content               string
	Description           string
	Status                TemplateStatus
	SMSSign               string
	SMSProviderTemplateId string
	Ctime                 int64
	Utime                 int64
}
