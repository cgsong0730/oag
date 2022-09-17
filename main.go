package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	_ "oag/docs"
	"strconv"
	"time"

	mongodb "oag/lib/mongodb"

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

// @title Open API Gateway
// @version 1.0
// @host localhost:5000
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
	e.GET("/users/:id", getUser)
	e.GET("/show", show)
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/forestPoint", forestPoint)

	e.Logger.Fatal(e.Start(":5000"))
}

// @Summary Get user
// @Description Get user's info
// @Accept json
// @Produce json
// @Param name path string true "name of the user"
// @Success 200 {object} User
// @Router /user/{name} [get]
func getUser(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

// @Summary Get show
// @Description Get show result
// @Accept json
// @Produce json
// @Param name path string true "name of the user"
// @Success 200 {string} string "ok"
// @Router /show [get]
func show(c echo.Context) error {
	team := c.QueryParam("team")
	member := c.QueryParam("member")
	return c.String(http.StatusOK, "team:"+team+", member:"+member)
}

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
				if cnt == 5 {
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

		returnStr = "count:" + strconv.Itoa(cnt) + "\n" + string(jsonData)
	}

	return c.String(http.StatusOK, returnStr)
}
