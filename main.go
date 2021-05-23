package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type ProvinceCityRegionModel struct {
	ProvinceCode      int    `gorm:"column:province_code" sql:"type:int(11)" json:"province_code"`
	ProvinceName      string `gorm:"column:province_name" sql:"type:varchar(128)" json:"province_name"`
	CityCode          int    `gorm:"column:city_code" sql:"type:int(11)" json:"city_code"`
	CityName          string `gorm:"column:city_name" sql:"type:varchar(128)" json:"city_name"`
	RegionCode        int    `gorm:"column:region_code" sql:"type:int(11)" json:"region_code"`
	RegionName        string `gorm:"column:region_name" sql:"type:varchar(128)" json:"region_name"`
}

type PublishRecord struct {
	//发布日期
	Date string `json:"date"`
	//链接
	Link string `json:"link"`
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

func main() {
	GetChinaAreaData()
}

func GetChinaAreaData() {

	publishRecords, err := GetPublishRecord()
	if err != nil {
		return
	}
	if len(publishRecords) == 0 {
		return
	}
	// 最新一条记录的更新日期
	// updatedAt = publishRecords[0].Date
	prefixUrl := publishRecords[0].Link

	provinces, err := GetProvinceUrlAndData(prefixUrl)
	if err != nil {
		log.Printf("GetProvinceUrlAndData err: %v", err)
		return
	}
	chinaAreaData, _ := json.Marshal(provinces)
	WriteWithIoutil("中国省市区数据", chinaAreaData)
}

// 将数据写入文件
func WriteWithIoutil(fileName string, data []byte) {
	if ioutil.WriteFile(fileName, data, 0644) == nil {
		fmt.Println("写入文件成功:")
	}
}

// 获取  http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/  这个页面的数据
// 根据ul[class='center_list_contlist'] 获取所有记录的更新日期及其链接地址
func GetPublishRecord() (publishRecords []PublishRecord, err error) {
	fetchUrl := "http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/"
	tempPublishRecords := make([]PublishRecord, 0)
	defer func() {
		if err == nil {
			publishRecords = tempPublishRecords
		}
	}()
	c := colly.NewCollector(colly.CacheDir("./缓存"))
	extensions.RandomUserAgent(c)
	// 设置gbk解码，防止乱码
	c.DetectCharset = true
	// 禁用 cookies
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	c.OnHTML("div[class='center'] div[class='center_list'] ul[class='center_list_contlist']", func(e *colly.HTMLElement) {
		e.ForEachWithBreak("ul li a ", func(i int, element *colly.HTMLElement) bool {
			hrefValue := element.Attr("href")
			if hrefValue == "" {
				err = fmt.Errorf("cant find herf value")
				return false
			}
			recordUrl := strings.ReplaceAll(hrefValue, filepath.Base(hrefValue), "")
			recordsUpdateTime := element.DOM.Find("span font[class='cont_tit02']").Text()
			if recordsUpdateTime == "" {
				err = fmt.Errorf("cant publish time value")
				return false
			}
			record := PublishRecord{
				Date: recordsUpdateTime,
				Link: recordUrl,
			}
			tempPublishRecords = append(tempPublishRecords, record)
			return true
		})
	})

	c.OnError(func(response *colly.Response, er error) {
		err = fmt.Errorf("visit %s OnError:%v", fetchUrl, er)
		return
	})
	if er := c.Visit(fetchUrl); er != nil {
		err = fmt.Errorf("visit %s error:%v", fetchUrl, er)
		return
	}
	return
}

// 获取所有省份对应的链接地址及省级数据
func GetProvinceUrlAndData(prefixUrl string) (provinces []Province, err error) {
	provs := make([]Province, 0)
	defer func() {
		if err == nil {
			provinces = provs
		}
	}()

	//Todo
	c := colly.NewCollector(colly.CacheDir("./缓存"))
	extensions.RandomUserAgent(c)
	// 设置gbk解码，防止乱码
	c.DetectCharset = true
	// 禁用 cookies
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	//省级列表
	c.OnHTML("tr[class='provincetr']", func(e *colly.HTMLElement) {
		//遍历每一行
		e.ForEachWithBreak("tr[class='provincetr'] > td", func(i int, item *colly.HTMLElement) bool {
			// 省份名称
			provinceName := item.ChildText("a")
			href := item.ChildAttr("a", "href")
			// 每个省份对应的链接地址, 最后一条td 里面没有 a 标签，排除这个td
			provinceHref := prefixUrl + href
			if href != "" && len(href) > 2 {
				provinceCode := href[:2]
				code, atoiErr := strconv.Atoi(provinceCode)
				if atoiErr != nil {
					return false
				}
				p := Province{
					Code:   code,
					Name:   provinceName,
					Link:   provinceHref,
					Cities: nil,
				}
				provs = append(provs, p)
			} else {
				err = errors.New("href 数据错误")
				return false
			}
			return true
		})
	})

	c.OnError(func(response *colly.Response, er error) {
		err = fmt.Errorf("visit %s OnError:%v", prefixUrl, er)
		return
	})

	err = c.Visit(prefixUrl)
	if err != nil {
		log.Printf("visit %s error: %v:\n", prefixUrl, err)
		return
	} else {
		log.Printf("visit %s\n", prefixUrl)
	}

	for i := 0; i < len(provs); i++ {
		cities, getCityErr := GetCityNameAndCode(prefixUrl, provs[i].Link)
		if getCityErr == nil {
			provs[i].Cities = cities
		} else {
			log.Printf("visit %s error: %v:\n", prefixUrl, err)
			err = getCityErr
			return
		}

		for j := 0; j < len(provs[i].Cities); j++ {

			if provs[i].Cities[j].Name == "东莞市" || provs[i].Cities[j].Name == "中山市" {
				towns, GetTownOfDonguanAndhongshanErr := GetTownOfDonguanAndhongshan(prefixUrl, (provs[i].Cities[j]).Link)
				if GetTownOfDonguanAndhongshanErr == nil {
					(provs[i].Cities[j]).Counties = towns
					continue
				} else {
					err = GetTownOfDonguanAndhongshanErr
					return
				}
			} else {
				counties, GetCountyNameAndCodeErr := GetCountyNameAndCode(prefixUrl, (provs[i].Cities[j]).Link)
				if GetCountyNameAndCodeErr == nil {
					(provs[i].Cities[j]).Counties = counties
				} else {
					err = GetCountyNameAndCodeErr
					return
				}
			}
		}
	}

	// 单独增加下面三个地区
	xianggang := Province{
		Code:   81,
		Name:   "香港特别行政区",
		Link:   "",
		Cities: nil,
	}
	aomeng := Province{
		Code:   82,
		Name:   "澳门特别行政区",
		Link:   "",
		Cities: nil,
	}
	taiwang := Province{
		Code:   71,
		Name:   "台湾省",
		Link:   "",
		Cities: nil,
	}

	provs = append(provs, xianggang, aomeng, taiwang)
	return
}

// 获取所有市的链接及市级数据
func GetCityNameAndCode(prefixUrl string, provinceUrl string) (cities []City, err error) {

	cts := make([]City, 0)
	defer func() {
		if err == nil {
			cities = cts
		}
	}()
	// Todo
	c := colly.NewCollector(colly.CacheDir("./缓存"))
	extensions.RandomUserAgent(c)
	// 设置gbk解码
	c.DetectCharset = true
	// 禁用 cookies
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	//市级列表
	c.OnHTML(".citytable tbody", func(e *colly.HTMLElement) {
		e.ForEachWithBreak("tr[class='citytr']", func(i int, item *colly.HTMLElement) bool {
			// 城市地址
			hrefValue, exists := item.DOM.Find("tr[class='citytr'] td >a").Attr("href")
			if !exists {
				log.Printf("hrefValue doesn't exists \n")
				return false
			}
			cityUrl := prefixUrl + hrefValue
			text := item.Text
			// 城市code
			if len(text) < 13 {
				log.Printf("获取市 len(text) < 13 ,数据有问题 \n")
				return false
			}
			cityCode := text[:4]
			code, atoiErr := strconv.Atoi(cityCode)
			if atoiErr != nil {
				log.Printf("strconv.Atoi(cityCode) error:%v", atoiErr)
				return false
			}
			// 城市名称
			cityName := text[12:]
			city := City{
				Code:     code,
				Name:     cityName,
				Link:     cityUrl,
				Counties: nil,
			}
			cts = append(cts, city)
			return true
		})
	})

	c.OnError(func(response *colly.Response, er error) {
		err = fmt.Errorf("visit %s OnError:%v", provinceUrl, er)
		return
	})

	if er := c.Visit(provinceUrl); er != nil {
		err = fmt.Errorf("visit %s error:%v", provinceUrl, er)
		return
	}
	return
}

// 存储所有区县的链接及其对应的名称和区划代码
func GetCountyNameAndCode(prefixUrl string, cityUrl string) (counties []County, err error) {

	couns := make([]County, 0)
	defer func() {
		if err == nil {
			counties = couns
		}
	}()
	c := colly.NewCollector(colly.CacheDir("./缓存"))
	extensions.RandomUserAgent(c)
	// 设置gbk解码
	c.DetectCharset = true
	// 禁用 cookies
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	//区县列表
	c.OnHTML(".countytable tbody", func(e *colly.HTMLElement) {
		//遍历每一行
		e.ForEachWithBreak("tr[class='countytr']", func(i int, item *colly.HTMLElement) bool {
			// 获取每个区县的url
			text := item.Text
			if len(text) < 13 {
				log.Println("获取区 len(text) < 13 ,数据有问题")
				return false
			}
			topTwo := text[:2]
			threeToFour := text[2:4]
			topSix := text[0:6]
			// 区县代码
			countyCode, _ := strconv.Atoi(topSix)
			// 区县的地址
			countyUrl := prefixUrl + topTwo + "/" + threeToFour + "/" + topSix + ".html"
			// 区县名称
			countyName := text[12:]
			county := County{
				Code: countyCode,
				Name: countyName,
				Link: countyUrl,
			}
			couns = append(couns, county)
			return true
		})
	})

	c.OnError(func(response *colly.Response, er error) {
		err = fmt.Errorf("visit %s OnError:%v", cityUrl, er)
		return
	})
	if er := c.Visit(cityUrl); er != nil {
		err = fmt.Errorf("visit %s error:%v", cityUrl, er)
		return
	}
	return
}

// 获取东莞市和中山市下属的所有镇名称和区划代码
func GetTownOfDonguanAndhongshan(prefixUrl string, cityUrl string) (counties []County, err error) {

	towns := make([]County, 0)
	defer func() {
		if err == nil {
			counties = towns
		}
	}()

	c := colly.NewCollector(colly.CacheDir("./缓存"))
	extensions.RandomUserAgent(c)
	// 设置gbk解码
	c.DetectCharset = true
	// 禁用 cookies
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	//镇列表
	c.OnHTML(".towntable tbody", func(e *colly.HTMLElement) {
		//遍历每一行
		e.ForEachWithBreak("tr[class='towntr']", func(i int, item *colly.HTMLElement) bool {
			// 获取每个镇的url
			text := item.Text
			if len(text) < 13 {
				log.Println("获取镇 len(text) < 13 ,数据有问题")
				return false
			}
			topTwo := text[:2]
			threeToFour := text[2:4]
			topNine := text[0:9]
			// 镇的地址
			townUrl := prefixUrl + topTwo + "/" + threeToFour + "/" + topNine + ".html"
			// 镇的code
			townCode, _ := strconv.Atoi(text[:6])
			// 镇名称
			townName := text[12:]
			// 镇级数据
			town := County{
				Code: townCode,
				Name: townName,
				Link: townUrl,
			}
			towns = append(towns, town)
			return true
		})
	})
	c.OnError(func(response *colly.Response, er error) {
		err = fmt.Errorf("visit %s OnError:%v", cityUrl, er)
		return
	})
	if er := c.Visit(cityUrl); er != nil {
		err = fmt.Errorf("visit %s error:%v", cityUrl, er)
		return
	}
	return
}

// 将抓取到的数据处理成对应数据库表的形式,省市区三级数据
func prepareData(provinces []Province) []ProvinceCityRegionModel {
	regions := make([]ProvinceCityRegionModel, 0)
	for _, p := range provinces {
		pro := ProvinceCityRegionModel{
			ProvinceCode:      p.Code,
			ProvinceName:      p.Name,
			CityCode:          0,
			CityName:          "",
			RegionCode:        0,
			RegionName:        "",
		}
		regions = append(regions, pro)
		for _, city := range p.Cities {
			cty := ProvinceCityRegionModel{
				ProvinceCode:      p.Code,
				ProvinceName:      p.Name,
				CityCode:          city.Code,
				CityName:          city.Name,
				RegionCode:        0,
				RegionName:        "",
			}
			regions = append(regions, cty)
			for _, county := range city.Counties {
				region := ProvinceCityRegionModel{
					ProvinceCode:      p.Code,
					ProvinceName:      p.Name,
					CityCode:          city.Code,
					CityName:          city.Name,
					RegionCode:        county.Code,
					RegionName:        county.Name,
				}
				regions = append(regions, region)
			}
		}
	}
	return regions
}
