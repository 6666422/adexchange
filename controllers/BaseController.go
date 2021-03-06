package controllers

import (
	"adexchange/engine"
	"adexchange/lib"
	m "adexchange/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/vmihailenco/msgpack.v2"
	//"strings"
)

type BaseController struct {
	beego.Controller
}

//Send log to the queue
//func SendLog(adRequest m.AdRequest, logType int) {

//engine.SendRequestLog(&adRequest, logType)
//b, err := msgpack.Marshal(adRequest)

//if err == nil {
//	c := lib.Pool.Get()
//	c.Do("lpush", getQueueName(logType), b)

//	defer c.Close()
//} else {

//	beego.Error(err.Error())
//}

//}

func getQueueName(logType int) string {
	prefix := beego.AppConfig.String("runmode") + "_"

	if logType == 1 {
		return prefix + "ADMUX_REQ"
	} else if logType == 2 {
		return prefix + "ADMUX_IMP"
	} else if logType == 3 {
		return prefix + "ADMUX_CLK"
	} else {
		return ""
	}
}

func GetClientIP(input *context.BeegoInput) string {
	//ips := input.Proxy()
	//if len(ips) > 0 && ips[0] != "" {
	//	return ips[0]
	//}
	//ip := strings.Split(input.Request.RemoteAddr, ":")
	//if len(ip) > 0 {
	//	return ip[0]
	//}
	return input.IP()
}

func SetCachedClkUrl(cacheKey string, clkUrl string) (err error) {
	c := lib.GetCachePool().Get()
	prefix := beego.AppConfig.String("runmode") + "_URL_"
	cachePeroid, _ := beego.AppConfig.Int("cache_clickurl_period")

	if _, err = c.Do("SETEX", prefix+cacheKey, cachePeroid, clkUrl); err != nil {
		beego.Error(err.Error())
	}

	//_, err = c.Do("EXPIRE", prefix+cacheKey, 3600)
	//if err != nil {
	//	beego.Error(err.Error())
	//}

	return
}

func GetCachedClkUrl(cacheKey string) (clkUrl string) {
	c := lib.GetCachePool().Get()
	prefix := beego.AppConfig.String("runmode") + "_URL_"
	beego.Debug(prefix + cacheKey)
	clkUrl, err := redis.String(c.Do("GET", prefix+cacheKey))

	if err != nil {
		beego.Error(err.Error())
	}

	return
}

func SetCachedAdResponse(cacheKey string, adResponse *m.AdResponse) {
	c := lib.GetCachePool().Get()
	prefix := beego.AppConfig.String("runmode") + "_"

	val, err := msgpack.Marshal(adResponse)

	if _, err = c.Do("SETEX", prefix+cacheKey, 3600, val); err != nil {
		beego.Error(err.Error())
	}

	//_, err = c.Do("EXPIRE", prefix+cacheKey, 3600)
	//if err != nil {
	//	beego.Error(err.Error())
	//}
}

func GetCachedAdResponse(cacheKey string) (adResponse *m.AdResponse) {
	c := lib.GetCachePool().Get()
	prefix := beego.AppConfig.String("runmode") + "_"

	v, err := c.Do("GET", prefix+cacheKey)
	if err != nil {
		beego.Error(err.Error())
		return nil
	}

	if v == nil {
		return
	}

	adResponse = new(m.AdResponse)
	switch t := v.(type) {
	case []byte:
		err = msgpack.Unmarshal(t, adResponse)
	default:
		err = msgpack.Unmarshal(t.([]byte), adResponse)
	}

	if err != nil {
		beego.Error(err.Error())
	}
	return
}

func GetCommonResponse(adResponse *m.AdResponse) (commonResponse m.CommonResponse) {
	commonResponse = adResponse.GenerateCommonResponse()

	if adResponse.Adunit != nil {
		if adResponse.Adunit.CreativeType == lib.CREATIVE_TYPE_HTML {
			cacheKey := adResponse.Did
			tplName := engine.GetPmpAdspaceTemplate(adResponse.AdspaceKey)

			url := beego.AppConfig.String("viewad_server") + "?id=" + cacheKey + "&tpl=" + tplName
			adResponse.AddClkTracking(adResponse.PmpClkTrackingUrl)
			commonResponse.SetHtmlCreativeUrl(url)
			SetCachedAdResponse(cacheKey, adResponse)
		} else {

			cacheKey := adResponse.ClickUrlCacheKey

			SetCachedClkUrl(cacheKey, adResponse.Adunit.ClickUrl)
			adResponse.Adunit.ClickUrl = adResponse.PmpClkTrackingUrl
		}
	}

	return

}

func ValidRequest(adRequest *m.AdRequest) (flg bool) {

	if adRequest.Os >= 2 {
		return
	}

	switch adRequest.Os {
	case 0:
		if len(adRequest.Imei) > 0 && len(adRequest.Aid) > 0 {
			flg = true
		}
		break
	case 1:
		if len(adRequest.Idfa) > 0 {
			flg = true
		}
		break
	case 2:
		if len(adRequest.Uid) > 0 {
			flg = true
		}

	default:
		break
	}
	return

}

//func GenerateBid(adRequest m.AdRequest) string {

//	return lib.GetMd5String(lib.GenerateBid(adRequest.AdspaceKey))

//}
