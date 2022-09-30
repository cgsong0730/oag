package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	_ "oag/docs"
	"time"

	mongodb "oag/lib/mongodb"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.mongodb.org/mongo-driver/bson"
)

// User struct
type User struct {
	name string `json:"name"`
	age  int    `json:"age"`
}

type OpenApi struct {
	params string
	count  int
}

// XML Data Architecture

type Response struct {
	XMLName        xml.Name `xml:"response"`
	ResponseHeader Header   `xml:"header"`
	ResponseBody   Body     `xml:"body"`
}

type Header struct {
	ResultCode string `xml:"resultCode"`
	ResultMsg  string `xml:"resultMsg"`
}

type Body struct {
	BodyItems Items `xml:"items"`
}

type Items struct {
	Item []Item `xml:"item"`
}

type Item struct {
	Analdate   string `xml:"analdate"`
	Area       string `xml:"area"`
	D1         string `xml:"d1"`
	D2         string `xml:"d2"`
	D3         string `xml:"d3"`
	D4         string `xml:"d4"`
	Domain     string `xml:"domain"`
	Maxi       string `xml:"maxi"`
	Meanavg    string `xml:"meanavg"`
	Mini       string `xml:"mini"`
	Regioncode string `xml:"regioncode"`
	Searchcd   string `xml:"searchcd"`
	Std        string `xml:"std"`
}

type Params struct {
	Key             string `json:key`
	PageNo          string `json:pageNo`
	NumOfRows       string `json:numOfRows`
	UType           string `json:_type`
	ExcludeForecast string `json:excludeForecast`
}

var client http.Client
var openApiList []OpenApi

// @title OAG(Open API Gateway)
// @version 1.0
// @host localhost:5000
// @description API Gateway for QoS Assurance of Open API Based Application
// @BasePath /
func main() {

	mongodb.Init()

	client = http.Client{
		Timeout: 60 * time.Second,
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/forestPoint", forestPoint)
	e.GET("/forestfire", forestfire)
	e.GET("/forestFires", forestFires)

	e.Logger.Fatal(e.Start(":5000"))
}

// @Summery Get forestfire
// @Description Get forestfire data.
// @Accept text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
// @Produce text/html
// @Success 200 {string} string "ok"
// @Router /forestfire [get]
func forestfire(c echo.Context) error {
	url := "http://forestfire.nifos.go.kr/main.action"
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	wrapper := html.Find("div.main_section_1")
	//subwrapper := wrapper.Find("div.thumb")
	items := wrapper.Find("li.item")

	resultStr := ""
	items.Each(func(idx int, sel *goquery.Selection) {
		if idx == 5 {
			img := sel.Find("img")
			src, exists := img.Attr("src")
			if exists {
				resultStr += "http://forestfire.nifos.go.kr" + src[2:]
			}
		}
	})
	nresp, err := client.Get(resultStr)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(nresp.Body)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(data))
}

// @Summery Get forestFires
// @Description Get forestFires data.
// @Accept text/html
// @Produce text/html
// @Success 200 {string} string "ok"
// @Router /forestFires [get]
func forestFires(c echo.Context) error {
	url := "https://d.kbs.co.kr/now/forestFires"
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	wrapper := html.Find("div.todayArea")
	items := wrapper.Find("span")

	resultStr := ""

	items.Each(func(idx int, sel *goquery.Selection) {
		if idx == 0 {
			resultStr += sel.Text()
		} else {
			resultStr += "," + sel.Text()
		}
	})

	return c.String(http.StatusOK, resultStr)
}

// @Summery Get forestPoint
// @Description Get forestPoint data.
// @Accept text/html
// @Produce text/html
// @Success 200 {string} string "ok"
// @Router /forestPoint [get]
func forestPoint(c echo.Context) error {
	url := "http://apis.data.go.kr/1400377/forestPoint/forestPointListGeongugSearch"
	key := "rsUsTZ5iD6e8vRno%2FpcvXuBJVdP4aYUwcONc0xcsPbGjHkCNpMnobuQb8ZcuAd7%2BrjqNlreQU9907P7OHW8N%2Fw%3D%3D&"
	pageNo := c.QueryParam("pageNo")
	numOfRows := c.QueryParam("numOfRows")
	_type := c.QueryParam("_type")
	excludeForecast := c.QueryParam("excludeForecast")

	query := url + "?serviceKey=" + key + "&pageNo=" + pageNo + "&numOfRows=" + numOfRows + "&_type=" + _type + "&excludeForecast=" + excludeForecast

	params := Params{
		Key:             key,
		PageNo:          pageNo,
		NumOfRows:       numOfRows,
		UType:           _type,
		ExcludeForecast: excludeForecast,
	}

	jsonParams, _ := json.Marshal(params)
	strParams := string(jsonParams)

	//cnt := 0
	isEqual := false
	var newOpenApi OpenApi

	dbDataStr := ""
	dbDatas := mongodb.ReadData("data", bson.M{"params": strParams}, bson.M{})
	dbResult := false
	returnStr := ""

	for _, dbData := range dbDatas {
		dbDataStr = dbData["data"].(string)
		dbResult = true
		returnStr = dbDataStr
	}

	if !dbResult {
		cnt := 0
		resp, err := client.Get(query)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var rs Response
		xmlerr := xml.Unmarshal(data, &rs)
		if xmlerr != nil {
			return xmlerr
		}

		jsonData, _ := json.MarshalIndent(rs, "", "\t")

		for i, child := range openApiList {
			if strParams == child.params {
				isEqual = true
				openApiList[i].count++
				cnt = child.count
				if cnt == 30 {
					mongodb.CreateData(strParams, string(jsonData))
				}
			}
		}

		if !isEqual {
			newOpenApi = OpenApi{
				params: strParams,
				count:  0,
			}
			openApiList = append(openApiList, newOpenApi)
		}

		//returnStr = "count:" + strconv.Itoa(cnt) + "\n" + string(jsonData)
		returnStr = string(jsonData)
	}

	return c.String(http.StatusOK, returnStr)
}
