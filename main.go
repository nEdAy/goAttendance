package main

import (
	"fmt"
	"github.com/apolloconfig/agollo/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"net/http"
)

type DayInfo struct {
	Code int `json:"code"`
	Type struct {
		Type int    `json:"type"`
		Name string `json:"name"`
		Week int    `json:"week"`
	} `json:"type"`
	Holiday struct {
		Holiday bool   `json:"holiday"`
		Name    string `json:"name"`
		Wage    int    `json:"wage"`
		Date    string `json:"date"`
		Rest    int    `json:"rest"`
	} `json:"holiday"`
}

type WechatToken struct {
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间，单位：秒。目前是7200秒之内的值。
	ErrCode     int    `json:"errcode"`      // 错误码
	ErrMsg      string `json:"errmsg"`       // 错误信息
}

type PhoneNumberResponse struct {
	ErrCode   int      `json:"errcode"` // 错误码
	ErrMsg    string   `json:"errmsg"`  // 错误提示信息
	PhoneInfo struct { // 用户手机号信息
		PhoneNumber     string   `json:"phoneNumber"`     // 用户绑定的手机号（国外手机号会有区号）
		PurePhoneNumber string   `json:"purePhoneNumber"` // 没有区号的手机号
		CountryCode     string   `json:"countryCode"`     // 区号
		Watermark       struct { // 数据水印
			AppId     string `json:"appid"`     // 小程序appid
			Timestamp int64  `json:"timestamp"` // 用户获取手机号操作的时间戳
		} `json:"watermark"`
	} `json:"phone_info"`
}

var appid, secret string

func initApolloClient() {
	client, err := agollo.Start()
	if err != nil {
		logrus.Error(err)
	} else {
		appid = client.GetStringValue("wechat-appid", "")
		secret = client.GetStringValue("wechat-secret", "")
		logrus.WithFields(logrus.Fields{
			"wechat-appid":  appid,
			"wechat-secret": secret,
		}).Debug("初始化Apollo配置成功")
	}
}

func initCron() {
	c := cron.New()
	c.AddFunc("@every 1s", func() {
		fmt.Println("tick every 1 second")
	})
	c.Start()
}

func main() {
	initApolloClient()
	initCron()
	router := gin.Default()
	router.Use(gin.Logger())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	/*获取指定日期的节假日信息
	接口地址：http://timor.tech/api/holiday/info/$date
	@params $date: 指定日期的字符串，格式 ‘2018-02-23’。可以省略，则默认服务器的当前时间。
	@return json: 如果不是节假日，holiday字段为null。
	{
		"code": 0,              // 0服务正常。-1服务出错
		"type": {
			"type": enum(0, 1, 2, 3), // 节假日类型，分别表示 工作日、周末、节日、调休。
			"name": "周六",         // 节假日类型中文名，可能值为 周一 至 周日、假期的名字、某某调休。
			"week": enum(1 - 7)    // 一周中的第几天。值为 1 - 7，分别表示 周一 至 周日。
		},
		"holiday": {
			"holiday": false,     // true表示是节假日，false表示是调休
			"name": "国庆前调休",  // 节假日的中文名。如果是调休，则是调休的中文名，例如'国庆前调休'
			"wage": 1,            // 薪资倍数，1表示是1倍工资
			"after": false,       // 只在调休下有该字段。true表示放完假后调休，false表示先调休再放假
			"target": '国庆节'     // 只在调休下有该字段。表示调休的节假日
		}
	}*/
	// attendance
	attendance := router.Group("/attendance")
	attendance.GET("/isWorkingDay/:day", func(c *gin.Context) {
		day := c.Param("day")
		client := resty.New()
		resp, err := client.R().
			SetPathParam("day", day).
			SetResult(&DayInfo{}).
			SetHeader("Accept", "application/json").
			Get("https://timor.tech/api/holiday/info/:day")
		if resp.IsSuccess() {
			dayInfo := resp.Result().(*DayInfo)
			if dayInfo.Code == 0 {
				c.JSON(http.StatusOK, dayInfo)
			} else {
				c.JSON(http.StatusOK, gin.H{"error": "服务出错"})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"error": err})
		}
	})
	// wechat
	wechat := router.Group("/wechat")
	wechat.GET("/userPhoneNumber", func(c *gin.Context) {
		code := c.Query("code")
		client := resty.New()
		resp, err := client.R().
			SetQueryParams(map[string]string{
				"appid":      appid,               // 小程序 appId
				"secret":     secret,              // 小程序 appSecret
				"grant_type": "client_credential", // 授权类型，此处只需填写 client_credential
			}).
			SetResult(&WechatToken{}).
			SetHeader("Accept", "application/json").
			Get("https://api.weixin.qq.com/cgi-bin/token")
		fmt.Println("  Status Code:", resp.StatusCode())
		if resp.IsSuccess() {
			token := resp.Result().(*WechatToken)
			if token.ErrCode == 0 {
				resp, err := client.R().
					SetQueryParam(
						"access_token", token.AccessToken, // 接口调用凭证
					).
					SetBody(map[string]string{
						"code": code, // 手机号获取凭证
					}).
					SetResult(&PhoneNumberResponse{}).
					SetHeader("Accept", "application/json").
					Post("https://api.weixin.qq.com/wxa/business/getuserphonenumber")
				if resp.IsSuccess() {
					phoneNumberResponse := resp.Result().(*PhoneNumberResponse)
					if phoneNumberResponse.ErrCode == 0 {
						c.JSON(http.StatusOK, gin.H{
							"phoneNumber":     phoneNumberResponse.PhoneInfo.PhoneNumber,
							"purePhoneNumber": phoneNumberResponse.PhoneInfo.PurePhoneNumber,
							"countryCode":     phoneNumberResponse.PhoneInfo.CountryCode,
							"watermark":       phoneNumberResponse.PhoneInfo.Watermark,
						})
					} else {
						c.JSON(http.StatusOK, gin.H{"error": phoneNumberResponse.ErrMsg})
					}
				} else {
					c.JSON(http.StatusOK, gin.H{"error": err})
				}
			} else {
				c.JSON(http.StatusOK, gin.H{"error": token.ErrMsg})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"error": err})
		}
	})
	router.RunTLS(":8443", "./tls/6881449_sunac.neday.cn.pem", "./tls/6881449_sunac.neday.cn.key")
}
