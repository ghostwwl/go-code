/*****************************************
 * FileName : txt2voice
 * Author   : ghostwwl
 * Note     : 测试百度语音的语音API呢
 *            将 中文 --> 语音 用的百度语音合成
 *            将 语音 --> 中文 用百度语音识别
 *****************************************/

package main

import (
	"encoding/json"
	"ghostlib"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	"errors"
	"encoding/base64"
	"fmt"
	"bytes"
	"strings"
)

const (
	BAIDU_VOICE_KEY  = "********************" // yuyin.baidu.com 自己申请去 反正百度或永久免费
	BAIDU_VOICE_CRET = "**************************"
)

const (
	TOKEN_API_URI = "https://openapi.baidu.com/oauth/2.0/token"
	TXT2VOICE_API_URI = "http://tsn.baidu.com/text2audio" // 语音合成
	VOICE2TXT_API_URI = "http://vop.baidu.com/server_api"  // 语音识别

	API_TIMEOUT = 120
)

type VoiceError struct {
	Err_no  int    `json:"err_no"`
	Err_msg string `json:"err_msg"`
	Sn      string `json:"sn"`
	Idx     int    `json:"idx"`
}

type TxtToVoice struct {
	client       *http.Client
	access_token string
	error_code map[float64]string
}

func NewTxtToVoice() *TxtToVoice {
	c := new(TxtToVoice)
	c.client = &http.Client{Timeout: API_TIMEOUT * time.Second}
	c.error_code = map[float64]string{
		500 : "不支持输入",
		501 : "输入参数不正确",
		502 : "token验证失败",
		503 : "合成后端错误",
		3300: "输入参数不正确",
		3301: "识别错误",
		3302: "验证失败",
		3303: "语音服务器后端问题",
		3304: "请求 GPS 过大，超过限额",
		3305: "产品线当前日请求数超过限额",
	}
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

/**
 * @param voicebyte  音频文件内容
 */
func (this *TxtToVoice) GetText(voicebyte []byte) (string, error) {
	// 真实使用时这里要判断过期时间 避免重复获取token	
	if this.access_token == "" {
		doflag, _ := this.getToken()
		if !doflag {
			return "", errors.New("获取access token 失败")
		}
	}

	/**
	 * 这里有个坑货 注意压缩格式 和采样率
	 * 如果格式对 但是采样率不对 会乱出结果
	 * 如果都格式不对 会返回 3300[输入参数错误]
	 */
	post_arg := map[string]interface{}{
		"format":  "wav",	// 压缩格式支持：pcm（不压缩）、wav、opus、speex、amr、x-flac
		"rate":  16000,	// 原始语音的录音格式目前只支持评测 8k/16k 采样率 16bit 位深的单声道语音
		"channel" : 1,
		"lan": "zh",	// 语种选择，中文=zh、粤语=ct、英文=en，不区分大小写，默认中文
		"token":  this.access_token,
		"cuid": "12:34:56:78",
		"len":  len(voicebyte),
		"speech":  base64.StdEncoding.EncodeToString(voicebyte),
	}
	post_json, err := json.Marshal(post_arg)
	if nil != err{
		return "", err
	}

	req, err := http.NewRequest("POST", VOICE2TXT_API_URI, bytes.NewReader(post_json))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, _ := this.client.Do(req)

	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	map_result := map[string]interface{}{}
	json.Unmarshal(data, &map_result)

	err_no := map_result["err_no"].(float64)
	if 0 == err_no {
		result := make([]string, 0)
		for _, r := range map_result["result"].([]interface{}) {
			result = append(result, ghostlib.ToString(r))
		}
		return strings.Join(result, ""), nil
	} else {
		return "", errors.New(this.error_code[err_no])
	}
}

/**
 * @param intxt 要合成语音的文字
 */
func (this *TxtToVoice) GetVoice(intxt string) (bool, []byte) {
	// 真实使用时这里要判断过期时间 避免重复获取token
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
		"per":  0,             // 发音人选择, 0为女声，1为男声，3为情感合成-度逍遥，4为情感合成-度丫丫，默认为普通女声
	}

	resp, _ := this.client.PostForm(TXT2VOICE_API_URI, ghostlib.InitPostData(post_arg))


	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	contentType := resp.Header.Get("Content-type")
	fmt.Printf("%v", contentType)
	switch contentType {
	case "audio/mp3":
		return true, data
	case "application/json":
		var errobj VoiceError
		if err := json.Unmarshal(data, &errobj); nil != err {
			return false, []byte(ghostlib.ToString(err))
			//panic(err.Error())
		} else {
			return false, []byte(errobj.Err_msg)
		}
	}

	return false, nil
}

func test_txt2voice() {
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
		fmt.Printf("\nlen:%v", len(result))
		err := ioutil.WriteFile("/data1/bd_voice.mp3", result, 0766)
		if err != nil {
			ghostlib.Msg("写入结果文件[/data1/bd_voice.mp3]出错", 3)
		}
	}
}

func test_voice2txt(){
	engine := NewTxtToVoice()
	// bd_voice.wav 输入的是 单声道16k采样
	r, err := ioutil.ReadFile("/data1/bd_voice.wav")
	if nil != err{
		panic(err)
	}

	txtresult, err := engine.GetText(r)
	if nil != err {
		fmt.Printf("\n%v\n", err)
	}
	fmt.Printf("%v\n", txtresult)
}


func main() {
	test_txt2voice()
	test_voice2txt()
}
