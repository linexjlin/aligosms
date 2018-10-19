package aligosms

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

func TestXMLParse(t *testing.T) {
	okResponse := `
		<?xml version='1.0' encoding='UTF-8'?>
		<SendSmsResponse>
			<Message>OK</Message>
			<RequestId>8E2A944B-1217-425F-9FB2-371E9329F223</RequestId>
			<BizId>990007539852874154^0</BizId>
			<Code>OK</Code>
		</SendSmsResponse>
	`
	errResponse := `
		<?xml version='1.0' encoding='UTF-8'?>
		<Error>
			<RequestId>37AC22B8-AA25-45DE-86A9-F2F7A160E3CB</RequestId>
			<HostId>dysmsapi.aliyuncs.com</HostId>
			<Code>SignatureDoesNotMatch</Code>
			<Message>
				Specified signature is not matched with our calculation. server string to sign is:GET&amp;%2F&amp;AccessKeyId%3DLTAIQvtsjCMaiWRJ%26Action%3DSendSms%26Format%3DXML%26PhoneNumbers%3D15123279507%26RegionId%3Dcn-hangzhou%26SignName%3DLeonzhangxf%26SignatureMethod%3DHMAC-SHA1%26SignatureNonce%3D1cc0b8eb-bc00-4edf-9c50-d726b02b0800%26SignatureVersion%3D1.0%26TemplateCode%3DSMS_133967651%26TemplateParam%3D%257B%2522code%2522%253A%2522654321%2522%257D%26Timestamp%3D2018-10-18T08%253A19%253A03Z%26Version%3D2017-05-25
			</Message>
			<Recommend>
				<![CDATA[https://error-center.aliyun.com/status/search?Keyword=SignatureDoesNotMatch&source=PopGw]]>
			</Recommend>
		</Error>
	`

	fmt.Printf("--okResponse %s \n", okResponse)
	fmt.Printf("--errResponse %s \n", errResponse)
	fmt.Println("")
	content := errResponse
	content = strings.Replace(content, xml.Header, "", 1)

	resp := SendSmsResponse{}
	err := xml.Unmarshal([]byte(content), &resp)
	if nil != err {
		panic(err)
	}

	fmt.Println(resp.String())
}

func TestSendMessage(t *testing.T) {
	// 发送短信测试
	accKeyId, accSecret, signName, tempCode, phone, outId, tempParam := getParams()

	sender := DefaultMessageSender
	sender.AccessKeyId = accKeyId
	sender.AccessSecret = accSecret
	res := sender.SendMsg(signName, tempCode, phone, outId, tempParam)
	fmt.Println(res.String())
}

func getParams() (accKeyId, accSecret, signName, tempCode, phone, outId string, tempParam map[string]string) {
	accessKeyId := "XXXXXXXXXXX"
	accessSecret := "XXXXXXXXXXX"
	// 签名
	signName = "XXXXXXXXXXX"
	// 短信模板
	// 嘿嘿，你的验证码是这个哦：${code}，快去验证吧。
	templateCode := "XXXXXXXXXXX"
	templateParam := map[string]string{} //"{\"code\":\"654321\"}"
	templateParam["code"] = "123123"
	phone = "XXXXXXXXXXX"
	outId = ""
	return accessKeyId, accessSecret, signName, templateCode, phone, outId, templateParam
}
