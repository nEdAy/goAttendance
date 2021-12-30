package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"net/http"
	"sunacAttendance/config"
)

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

func main() {
	router := gin.Default()
	router.Use(gin.Logger())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// wechat
	wechat := router.Group("/wechat")
	wechat.GET("/userPhoneNumber", func(c *gin.Context) {
		code := c.Param("code")
		client := resty.New()
		client.SetDebug(true)
		resp, err := client.R().
			SetQueryParams(map[string]string{
				"appid":      "wxa1afc52994e69d4c",   // 小程序 appId
				"secret":     config.WechatSecretKey, // 小程序 appSecret
				"grant_type": "client_credential",    // 授权类型，此处只需填写 client_credential
			}).
			SetResult(&WechatToken{}).
			SetHeader("Accept", "application/json").
			Get("https://api.weixin.qq.com/cgi-bin/token")
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
