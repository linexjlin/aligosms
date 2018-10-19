package aligosms

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/satori/go.uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//参数Key列表
//The parameters key list.
const (
	//系统参数KEY
	//System parameter key.
	KeySignatureMethod  = "SignatureMethod"
	KeySignatureNonce   = "SignatureNonce"
	KeyAccessKeyId      = "AccessKeyId"
	KeySignatureVersion = "SignatureVersion"
	KeyTimestamp        = "Timestamp"
	KeyFormat           = "Format"

	//业务API参数KEY
	//Business API parameter key.
	KeyAction        = "Action"
	KeyVersion       = "Version"
	KeyRegionId      = "RegionId"
	KeyPhoneNumbers  = "PhoneNumbers"
	KeySignName      = "SignName"
	KeyTemplateParam = "TemplateParam"
	KeyTemplateCode  = "TemplateCode"
	KeyOutId         = "OutId"

	//签名参数KEY
	//Signature parameter key.
	KeySignature = "Signature"
)

//参数默认值列表
//Default value of parameters.
const (
	DefaultSignatureMethod  = "HMAC-SHA1"
	DefaultSignatureVersion = "1.0"
	DefaultFormat           = "XML"
	DefaultAction           = "SendSMS"
	DefaultVersion          = "2017-05-25"
	DefaultRegionId         = "cn-hangzhou"
	DefaultDomain           = "dysmsapi.aliyuncs.com"
	DefaultProtocol         = "http"
)

var (
	ParamsMarshallError     = errors.New("params marshall error")
	GetTimeLocationError    = errors.New("get time location error")
	ResponseUnMarshallError = errors.New("response unMarshall error")
	ResponseReadError       = errors.New("response read error")
	HttpRequestError        = errors.New("request error")
)

// 发送短信响应
// 正常时，为"OK" http状态码为200
// 出现异常时将出现以下值：HostId，Recommend
type SendSmsResponse struct {
	XMLNameSendSMSResponse xml.Name `xml:"SendSMSResponse"`
	XMLNameError           xml.Name `xml:"Error"`
	RequestId              string   `xml:"RequestId"`
	BizId                  string   `xml:"BizId"`
	Code                   string   `xml:"Code"`
	Message                string   `xml:"Message"`
	HostId                 string   `xml:"HostId"`
	Recommend              string   `xml:"Recommend"`
}

func (resp *SendSmsResponse) String() string {
	str := "SendSmsResponse = { " + " \n	RequestId: " + resp.RequestId
	str += " \n	BizId: " + resp.BizId
	str += " \n	Code: " + resp.Code
	str += " \n	Message: " + resp.Message
	str += " \n	HostId: " + resp.HostId
	str += " \n	Recommend: " + resp.Recommend
	str += " \n}"
	return str
}

// 消息发送器
type MessageSender struct {
	Protocol     string
	Domain       string
	RegionId     string
	AccessKeyId  string
	AccessSecret string
}

// 默认短信发送器
// 需要设置 AccessKeyId 和 AccessSecret 参数
var DefaultMessageSender = &MessageSender{
	Protocol: DefaultProtocol,
	Domain:   DefaultDomain,
	RegionId: DefaultRegionId,
}

// 创建短信发送器
func NewMessageSender() MessageSender {
	return MessageSender{
		Protocol: DefaultProtocol,
		Domain:   DefaultDomain,
		RegionId: DefaultRegionId,
	}
}

// 调用发送短信
// 需要设置 AccessKeyId 和 AccessSecret 参数
func (sender *MessageSender) SendMsg(signName, tempCode, phones, outId string,
	params map[string]string) SendSmsResponse {
	return SendMsg(sender.Protocol, sender.Domain, sender.RegionId,
		sender.AccessKeyId, sender.AccessSecret,
		signName, tempCode, phones, outId, params)
}

// 使用部分默认配置进行短信请求发送
// @param phones	String	必须	 15123279507 短信接收号码
// 	 支持以逗号分隔的形式进行批量调用，批量上限为1000个手机号码,
// 	 批量调用相对于单条调用及时性稍有延迟,验证码类型的短信推荐使用单条调用的方式；
// 	 发送国际/港澳台消息时，接收号码格式为：国际区号+号码，如“85200000000”
func SendMsg(protocol, domain, regionId, accKey, accSecret, signName, tempCode, phones, outId string,
	params map[string]string) SendSmsResponse {
	//1.封装参数
	//2.参数排序
	paramsMap := resolveParamsMap(regionId, accKey, phones, signName, tempCode, outId, params)

	//3.构造待签名的请求串
	//3.1 特殊URL编码
	urlParamsStr := resolveParamsStr(paramsMap)

	//3.2 按POP的签名规则拼接成最终的待签名串，规则如下：
	//* HTTPMethod + “&” + specialUrlEncode(“/”) + ”&” + specialUrlEncode(sortedQueryString)
	stringToSign := resolveStringToSign(urlParamsStr)

	//4.签名
	//签名采用HmacSHA1算法 + Base64，编码采用：UTF-8
	//5.增加签名结果到请求参数中，发送请求
	//注意：签名也要做特殊URL编码
	sign := Sign(accSecret, stringToSign)
	requestUri := protocol + "://" + domain + "/?"
	requestUri += urlParamsStr + "&" + SpecialUrlEncode("Signature") + "=" + sign

	//6.请求并解析响应
	resp := requestSendMsg(requestUri)
	return resp
}

// 请求短信发送并返回响应
func requestSendMsg(requestUri string) SendSmsResponse {
	httpClient := http.DefaultClient
	response, err := httpClient.Get(requestUri)
	if nil != err {
		panic(HttpRequestError)
	}
	readCloser := response.Body
	resp := resolveResp(readCloser)
	return resp

}

// 处理请求响应
func resolveResp(readCloser io.ReadCloser) SendSmsResponse {
	buffer := []byte{0}
	content := ""
	for num, err := readCloser.Read(buffer); num > 0; num, err = readCloser.Read(buffer) {
		if io.EOF == err {
			content += string(buffer)
			break
		}
		if nil != err {
			panic(ResponseReadError)
		}
		content += string(buffer)
	}
	readCloser.Close()
	if "" != content {
		content = strings.Replace(content, xml.Header, "", 1)
	}
	resp := SendSmsResponse{}
	err := xml.Unmarshal([]byte(content), &resp)
	if nil != err {
		panic(ResponseUnMarshallError)
	}
	return resp
}

// 生成待签名字符串
// 按POP的签名规则拼接成最终的待签名串，规则如下：
// - HTTPMethod + “&” + specialUrlEncode(“/”) + ”&” + specialUrlEncode(sortedQueryString)
func resolveStringToSign(urlParamsStr string) string {
	stringToSign := "GET&"
	stringToSign += SpecialUrlEncode("/") + "&"
	stringToSign += SpecialUrlEncode(urlParamsStr)
	return stringToSign
}

// 将参数封装，并返回字母有序排列映射
func resolveParamsMap(regionId, accKey, phones, signName, tempCode, outId string, params map[string]string) *treemap.Map {
	//1.请求参数
	//2.根据参数Key排序（顺序）字母顺序
	paramsMap := treemap.NewWithStringComparator()
	//请求参数包括系统参数和业务参数，不要遗漏
	//请求参数中不允许出现以Signature为key的参数。参考代码如下

	//1.1 系统参数
	paramsMap.Put(KeySignatureMethod, DefaultSignatureMethod)
	paramsMap.Put(KeySignatureNonce, uuid.Must(uuid.NewV4()).String())
	paramsMap.Put(KeyAccessKeyId, accKey)
	paramsMap.Put(KeySignatureVersion, DefaultSignatureVersion)
	paramsMap.Put(KeyFormat, DefaultFormat)
	location, err := time.LoadLocation("GMT")
	if nil != err {
		panic(GetTimeLocationError)
	}
	paramsMap.Put(KeyTimestamp, time.Now().In(location).Format("2006-01-02T15:04:05Z"))

	//1.2 业务API参数
	paramsMap.Put(KeyAction, DefaultAction)
	paramsMap.Put(KeyVersion, DefaultVersion)

	if "" == strings.Trim(regionId, " ") {
		paramsMap.Put(KeyRegionId, DefaultRegionId)
	} else {
		paramsMap.Put(KeyRegionId, regionId)
	}
	paramsMap.Put(KeyPhoneNumbers, phones)
	paramsMap.Put(KeySignName, signName)
	paramsMap.Put(KeyTemplateCode, tempCode)

	//模板变量，可选
	if nil != params {
		paramsBytes, err := json.Marshal(params)
		if nil != err {
			panic(ParamsMarshallError)
		}
		templateParam := string(paramsBytes)
		paramsMap.Put(KeyTemplateParam, templateParam)
	}

	//OutId	String	可选 部流水扩展字段
	if "" != outId {
		paramsMap.Put(KeyOutId, outId)
	}

	//1.3 去除签名关键字Key
	if _, hasSignature := paramsMap.Get(KeySignature); hasSignature {
		paramsMap.Remove(KeySignature)
	}
	return paramsMap
}

// 拼接生成URL请求参数字符
// @param paramsMap 参数映射（按照字母顺序排列）
// @return URL请求参数列表
func resolveParamsStr(paramsMap *treemap.Map) string {
	urlParamsStr := ""
	iterator := paramsMap.Iterator()
	for iterator.Next() {
		urlParamsStr += "&" + SpecialUrlEncode(iterator.Key().(string))
		urlParamsStr += "=" + SpecialUrlEncode(iterator.Value().(string))
	}
	urlParamsStr = strings.Replace(urlParamsStr, "&", "", 1)
	return urlParamsStr
}

// 特殊URL编码
// 这个是POP特殊的一种规则，即在一般的URLEncode后再增加三种字符替换：
// 加号（+）替换成 %20、星号（*）替换成 %2A、%7E 替换回波浪号（~）
// @param value 待编码字符串
// @return POP编码完成字符串
func SpecialUrlEncode(value string) (urlEncodeStr string) {
	urlEncodeStr = url.QueryEscape(value)
	urlEncodeStr = strings.Replace(urlEncodeStr, "+", "%20", -1)
	urlEncodeStr = strings.Replace(urlEncodeStr, "*", "%2A", -1)
	urlEncodeStr = strings.Replace(urlEncodeStr, "%7E", "~", -1)
	return urlEncodeStr
}

// 签名采用HmacSHA1算法 + Base64，编码采用：UTF-8
// 增加签名结果到请求参数中，发送请求
// 注意：签名也要做特殊URL编码
// @param accessSecret：你的AccessKeyId对应的秘钥AccessSecret，
// 			特别说明：POP要求需要后面多加一个“&”字符，即accessSecret + “&”
// @param stringToSign：即第三步生成的待签名请求串
// @return 签名
func Sign(accessSecret, stringToSign string) (sign string) {
	mac := hmac.New(sha1.New, []byte(accessSecret+"&"))
	mac.Write([]byte(stringToSign))
	signData := mac.Sum(nil)
	sign = base64.StdEncoding.EncodeToString(signData)
	return SpecialUrlEncode(sign)
}
