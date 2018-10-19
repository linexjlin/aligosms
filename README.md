aligosms
===============================

An Aliyun Short Message SDK by golang.

阿里云短信服务 Go语言SDK

## How to use

### Way 1

```go
    import "github.com/leonzhangxf/aligosms"

    sender := aligosms.DefaultMessageSender
	sender.AccessKeyId = accKeyId
	sender.AccessSecret = accSecret
	res := sender.SendMsg(signName, tempCode, phone, outId, tempParam)
	fmt.Println(res.String())
```

### Way 2

```go
    import "github.com/leonzhangxf/aligosms"
    

    aligosms.SendMsg(protocol, domain, regionId, accKey, 
    	    accSecret, signName, tempCode, phones, outId,
        	params)

```