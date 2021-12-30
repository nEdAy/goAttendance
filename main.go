package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"net/http"
	"sunacAttendance/config"
)

type WechatToken struct {
	access_token string // 获取到的凭证
	expires_in   int    // 凭证有效时间，单位：秒。目前是7200秒之内的值。
	errcode      int    // 错误码
	errmsg       string // 错误信息
}

type PhoneNumberResponse struct {
	errcode    int      // 错误码
	errmsg     string   // 错误提示信息
	phone_info struct { // 用户手机号信息
		phoneNumber     string   // 用户绑定的手机号（国外手机号会有区号）
		purePhoneNumber string   // 没有区号的手机号
		countryCode     string   // 区号
		watermark       struct { // 数据水印
			appid     string // 小程序appid
			timestamp int64  // 用户获取手机号操作的时间戳
		}
	}
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
		resp, err := client.R().
			SetQueryParams(map[string]string{
				"appid":      "wxa1afc52994e69d4c",   // 小程序 appId
				"secret":     config.WechatSecretKey, // 小程序 appSecret
				"grant_type": "client_credential",    // 授权类型，此处只需填写 authorization_code
			}).
			SetResult(&WechatToken{}).
			SetHeader("Accept", "application/json").
			Get("https://api.weixin.qq.com/cgi-bin/token")
		if resp.IsSuccess() {
			token := resp.Result().(*WechatToken)
			if token.errcode == 0 {
				resp, err := client.R().
					SetQueryParam(
						"access_token", token.access_token, // 接口调用凭证
					).
					SetBody(map[string]string{
						"code": code, // 手机号获取凭证
					}).
					SetResult(&PhoneNumberResponse{}).
					SetHeader("Accept", "application/json").
					Post("https://api.weixin.qq.com/wxa/business/getuserphonenumber")
				if resp.IsSuccess() {
					phoneNumberResponse := resp.Result().(*PhoneNumberResponse)
					if phoneNumberResponse.errcode == 0 {
						c.JSON(http.StatusOK, gin.H{
							"phoneNumber":     phoneNumberResponse.phone_info.phoneNumber,
							"purePhoneNumber": phoneNumberResponse.phone_info.purePhoneNumber,
							"countryCode":     phoneNumberResponse.phone_info.countryCode,
							"watermark":       phoneNumberResponse.phone_info.watermark,
						})
					} else {
						c.JSON(http.StatusOK, gin.H{"error": phoneNumberResponse.errmsg})
					}
				} else {
					c.JSON(http.StatusOK, gin.H{"error": err})
				}
			} else {
				c.JSON(http.StatusOK, gin.H{"error": token.errmsg})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"error": err})
		}
	})
	router.RunTLS(":8443", "./tls/6881449_sunac.neday.cn.pem", "./tls/6881449_sunac.neday.cn.key")
}
