package models

// 此处自行导入 gorm
import "gorm"

const TableName = "province_city_region"

type ProvinceCityRegionModel struct {
	ID                int    `gorm:"column:id" sql:"type:int(11)" json:"id"`
	ProvinceCode      int    `gorm:"column:province_code" sql:"type:int(11)" json:"province_code"`
	ProvinceName      string `gorm:"column:province_name" sql:"type:varchar(128)" json:"province_name"`
	ProvinceNamePy    string `gorm:"column:province_name_py" sql:"type:varchar(128)" json:"province_name_py"`
	CityCode          int    `gorm:"column:city_code" sql:"type:int(11)" json:"city_code"`
	CityName          string `gorm:"column:city_name" sql:"type:varchar(128)" json:"city_name"`
	CityNamePy        string `gorm:"column:city_name_py" sql:"type:varchar(128)" json:"city_name_py"`
	RegionCode        int    `gorm:"column:region_code" sql:"type:int(11)" json:"region_code"`
	RegionName        string `gorm:"column:region_name" sql:"type:varchar(128)" json:"region_name"`
	RegionNamePy      string `gorm:"column:region_name_py" sql:"type:varchar(128)" json:"region_name_py"`
	CityCodeTelephone string `gorm:"column:city_code_telephone" sql:"type:varchar(8)" json:"city_code_telephone"`
	Area              string `gorm:"column:area" sql:"type:varchar(64)" json:"area"`
}

type Province struct {
	Code   int    `json:"code"`
	Name   string `json:"name"`
	Link   string `json:"-"`
	Cities []City `json:"cities"`
}

type City struct {
	Code     int      `json:"code"`
	Name     string   `json:"name"`
	Link     string   `json:"-"`
	Counties []County `json:"counties"`
}

type County struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	Link string `json:"-"`
}

type ProvinceCityRegionModelList []ProvinceCityRegionModel

type DB struct {
	*gorm.DB
}

var db *DB

// 获取所有的不重复的省code
func (list *ProvinceCityRegionModelList) GetAllProvince() {
	db.Table(TableName).Select("distinct(province_code)").Find(&list)
}

func (list *ProvinceCityRegionModelList) GetAllOrderAsc() error {
	err := db.Table(TABLE_NAME).Order("province_code, city_code, region_code").Find(list).Error
	return err
}

// 根据省code获取该省下属的所有不重复的市
func (list *ProvinceCityRegionModelList) GetCityListOfSingleProvince(provinceCode int) {
	db.Table(TableName).Select("distinct(city_code)").Where("province_code = ?", provinceCode).Where("city_code != ?", 0).Find(&list)
}

// 判断该市下面是否有区级数据
func (p *ProvinceCityRegionModel) HasCounty(cityCode int) bool {
	var count int
	db.Table(TableName).Where("city_code = ?", cityCode).Where("region_code != ?", 0).Count(&count)
	return count >= 1
}
