/*****************************************
 * FileName : txt2voice.go
 * Author   : ghostwwl
 * Note     : 测试百度语音的语音合成API呢
 *            将 中文 --> 语音 用的百度TTS呢
 *****************************************/

package main

import (
	//	"fmt"
	"ghostlib"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	BAIDU_VOICE_KEY  = "******************"   // yuyin.baidu.com 自己申请去 反正百度或永久免费
	BAIDU_VOICE_CRET = "*******************"
)

const (
	TOKEN_API_URI = "https://openapi.baidu.com/oauth/2.0/token"
	VOICE_API_URI = "http://tsn.baidu.com/text2audio"

	API_TIMEOUT = 120
)

type TxtToVoice struct {
	client       *http.Client
	access_token string
}

func NewTxtToVoice() *TxtToVoice {
	c := new(TxtToVoice)
	c.client = &http.Client{Timeout: API_TIMEOUT * time.Second}
	return c
}

func (this *TxtToVoice) getToken() (bool, string) {
	post_arg := map[string]interface{}{
		"client_id":     BAIDU_VOICE_KEY,
		"client_secret": BAIDU_VOICE_CRET,
		"grant_type":    "client_credentials",
	}

	resp, _ := this.client.PostForm(TOKEN_API_URI, ghostlib.InitPostData(post_arg))
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	json_result, err := simplejson.NewJson(data)
	if err != nil {
		panic(err)
	}

	map_result := make(map[string]interface{})
	map_result, _ = json_result.Map()
	access_token, ok := map_result["access_token"]

	if ok {
		this.access_token = ghostlib.ToString(access_token)
		return true, this.access_token
	}
	return false, ""
}

func (this *TxtToVoice) GetVoice(intxt string) (bool, []byte) {
	if this.access_token == "" {
		doflag, _ := this.getToken()
		if !doflag {
			return false, []byte("获取access token 失败")
		}
	}

	post_arg := map[string]interface{}{
		"tex":  intxt,
		"lan":  "zh",
		"tok":  this.access_token,
		"ctp":  1,
		"cuid": "12:34:56:78", // 用户唯一标识,用来区分用户,web 端参考填写机器 mac地址或 imei 码,长度为 60 以内
		"spd":  3,             // 语速,取值 0-9,默认为 5
		"pit":  3,             // 音调,取值 0-9,默认为 5
		"vol":  5,             // 音量,取值 0-9,默认为 5
		"per":  0,             // 发音人选择,取值 0-1 ;0 为女声,1 为男声,默认为女声
	}

	resp, _ := this.client.PostForm(VOICE_API_URI, ghostlib.InitPostData(post_arg))

	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	if resp.Header["Content-Type"][0] == "application/json" {
		return false, data
	} else {
		return true, data
	}

}

func main() {
	T := `
晚上跑完步有点口渴，到路边摊买橘子，挑了几个卖相好的。
老板又拿起几个有斑点的说：“帅哥，这种长得不好看的其实更甜。”
我颇有感悟的说：“是因为橘子觉得自己长的不好看，所以努力让自己变得更甜吗？”
老板微微一愣道：“不是，我想早点卖完回家。”
	`
	engine := NewTxtToVoice()
	flag, result := engine.GetVoice(T)
	if !flag {
		ghostlib.Msg(string(result), 3)
	} else {
		err := ioutil.WriteFile("/data0/bd_voice.mp3", result, 766)
		if err != nil {
			ghostlib.Msg("写入结果文件[/data0/bd_voice.mp3]出错", 3)
		}
	}

}
