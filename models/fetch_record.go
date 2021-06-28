package models

// 此处自行导入 gorm
import "gorm"

// 存储国家统计局数据上传到oss的链接记录
type FetchRecord struct {
	Id int `gorm:"column:id" form:"id"`
	// 数据库该条记录的更新时间
	UpdateTime string `gorm:"update_time" form:"update_time"`
	// 国家统计局最新一条记录的发布时间
	UpdateAt string `gorm:"column:update_at" form:"update_at"`
	// 省市区数据oss下载链接
	DownUrl string `gorm:"column:down_url" form:"down_url"`
}

type CsvFile struct {
	Name string
	Data []byte
}

// 国家统计局每条记录对应的发布日期与链接
type PublishRecord struct {
	//发布日期
	Date string `json:"date"`
	//链接
	Link string `json:"link"`
}

func (f *FetchRecord) TableName() string {
	return "fetch_record"
}

func (f *FetchRecord) Create(db *gorm.DB) error {
	return db.Table(f.TableName()).Create(f).Error
}

func (f *FetchRecord) GetNewestFetchRecord(db *gorm.DB) (err error) {
	err = db.Table(f.TableName()).Order("id desc").Limit(1).Find(f).Error
	return
}
