// Artxun Wallets SDK, (C) 2016 ghostwwl@gmail.com.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package ArtxunWallets

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"ghostlib"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/bitly/go-simplejson"
	"github.com/bradfitz/gomemcache/memcache"
)

const (
	formatTime     = "15:04:05"
	formatDate     = "2006-01-02"
	formatDateTime = "2006-01-02 15:04:05"
)

const (
	FINANCE_TOKEN_API = "http://pay.artxun.com/token.php"
	FINANCE_CORE_API  = "http://pay.artxun.com/service.php"

	CLIENT_ID     = "**********************"
	CLIENT_SECRET = "**********************"

	// 充值密码       充值订单密码 用户申请充值 、充值订单查询接口 参与 hash签名
	pwd_charge = "**********************"
	// 提现密码     提现订单密码 用户申请提现、提现订单查询接口 参与 hash签名
	pwd_withdraw = "**********************"
	// 交易密码
	pwd_trade = "**********************"

	USER_TOKEN_API = "http://oauth.artxun.com/token.php"
	USER_CORE_API  = "http://oauth.artxun.com/scope.php"
)

const (
	CACHE_SERVER = "localhost:11211"
)

//---------------- end const -------------------

type FinanceCache struct {
	memc *memcache.Client
}

func NewFinanceCache() *FinanceCache {
	c := new(FinanceCache)
	c.memc = memcache.New(CACHE_SERVER)
	return c
}

func (this *FinanceCache) set_cache(k, v string, exptime int32) (bool, error) {
	if exptime > 0 {
		if err := this.memc.Set(&memcache.Item{Key: k, Value: []byte(v), Expiration: exptime}); err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		if err := this.memc.Set(&memcache.Item{Key: k, Value: []byte(v)}); err != nil {
			return false, err
		}
		return true, nil
	}
}

func (this *FinanceCache) get_cache(k string) (string, error) {
	if it, err := this.memc.Get(k); err != nil {
		return "", err
	} else {
		return string(it.Value), nil
	}
}

func (this *FinanceCache) del_cache(k string) (bool, error) {
	err := this.memc.Delete(k)
	if err != nil {
		return false, err
	}
	return true, err
}

//------------------------ end finance cache --------------------

type financeuser struct {
	artxun_openid string
	artxun_uid    uint32
}

func NewFinanceUser(openid string, uid uint32) *financeuser {
	return &financeuser{
		artxun_openid: openid,
		artxun_uid:    uid,
	}
}

type FinanceApi struct {
	access_token string
	scope        string
	cookies      []*http.Cookie
	client       *http.Client
}

type UserApi struct {
	access_token string
	scop         string
	cookies      []*http.Cookie
	client       *http.Client
}

func NewFinanceApi() *FinanceApi {
	f := new(FinanceApi)
	f.client = &http.Client{}
	return f
}

func NewUserApi() *UserApi {
	u := new(UserApi)
	u.client = &http.Client{}
	return u

}

func (this *UserApi) getToken() (bool, string) {
	fcache := NewFinanceCache()
	utoken_key := "__c2cmall_user_token"
	user_token, _ := fcache.get_cache("__c2cmall_user_token")

	if user_token != "" {
		this.access_token = user_token
		return true, this.access_token
	} else {
		post_arg := map[string]interface{}{
			"client_id":     CLIENT_ID,
			"client_secret": CLIENT_SECRET,
			"grant_type":    "client_credentials",
		}

		resp, _ := this.client.PostForm(USER_TOKEN_API, ghostlib.InitPostData(post_arg))
		data, _ := ioutil.ReadAll(resp.Body)

		//fmt.Printf("U:%s\n", string(data))
		json_result, err := simplejson.NewJson(data)
		if err != nil {
			panic(err)
		}

		map_result := make(map[string]interface{})
		map_result, _ = json_result.Map()
		access_token, ok := map_result["access_token"]
		error_description, has_error := map_result["error_description"]

		if ok {
			this.access_token = ghostlib.ToString(access_token)
			fcache.set_cache(utoken_key, this.access_token, 600)
			return true, this.access_token
		}

		if has_error {
			return false, ghostlib.ToString(error_description)
		}

		return false, ""
	}
}

func (this *UserApi) doUserApi(post_arg map[string]interface{}) *map[string]interface{} {
	if this.access_token == "" {
		doflag, _ := this.getToken()
		if !doflag {
			panic("获取用户接口TOKEN失败")
		}
	}

	post_arg["access_token"] = this.access_token

	resp, _ := this.client.PostForm(USER_CORE_API, ghostlib.InitPostData(post_arg))
	data, _ := ioutil.ReadAll(resp.Body)

	json_result, err := simplejson.NewJson(data)
	if err != nil {
		panic(err)
	}

	map_result := make(map[string]interface{})
	map_result, _ = json_result.Map()

	return &map_result

}

func (this *UserApi) UserHas(username string) *map[string]interface{} {
	post_arg := map[string]interface{}{
		"scope":    "user_has",
		"username": username,
	}

	return this.doUserApi(post_arg)
}

func (this *UserApi) RegUser(username, password string) *map[string]interface{} {
	post_arg := map[string]interface{}{
		"scope":    "user_reg",
		"username": username,
		"password": password,
	}

	return this.doUserApi(post_arg)

}

func (this *UserApi) getUserInfo(artxun_openid string) *map[string]interface{} {
	post_arg := map[string]interface{}{
		"scope":         "user_info",
		"artxun_openid": artxun_openid,
	}

	return this.doUserApi(post_arg)
}

//-----------------------------------------------------------------------------
func (this *FinanceApi) getToken() (bool, string) {
	post_arg := map[string]interface{}{
		"client_id":     CLIENT_ID,
		"client_secret": CLIENT_SECRET,
		"grant_type":    "client_credentials",
	}

	resp, _ := this.client.PostForm(FINANCE_TOKEN_API, ghostlib.InitPostData(post_arg))
	data, _ := ioutil.ReadAll(resp.Body)

	json_result, err := simplejson.NewJson(data)
	if err != nil {
		panic(err)
	}

	map_result := make(map[string]interface{})
	map_result, _ = json_result.Map()
	access_token, ok := map_result["access_token"]
	error_description, has_error := map_result["error_description"]

	if ok {
		this.access_token = ghostlib.ToString(access_token)
		return true, this.access_token

	}

	if has_error {
		return false, ghostlib.ToString(error_description)
	}

	return false, ""
}

func (this *FinanceApi) doFinanceApi(post_arg map[string]interface{}) *map[string]interface{} {
	if this.access_token == "" {
		doflag, _ := this.getToken()
		if !doflag {
			panic("获取财务接口TOKEN失败")
		}
	}

	post_arg["access_token"] = this.access_token

	resp, _ := this.client.PostForm(FINANCE_CORE_API, ghostlib.InitPostData(post_arg))
	data, _ := ioutil.ReadAll(resp.Body)

	json_result, err := simplejson.NewJson(data)
	if err != nil {
		panic(err)
	}

	map_result := make(map[string]interface{})
	map_result, _ = json_result.Map()

	rsign, ok := (map_result)["sign"]
	if ok {
		result_sing_str, _ := this.genSignString(map_result, pwd_charge)
		verify_flag, err := this.verifyArtxunSign(ghostlib.ToString(rsign), result_sing_str)
		if err != nil {
			panic(err)
		}

		if verify_flag {
			delete(map_result, "sign")
			return &map_result
		} else {
			panic("返回结果签名验证失败")
		}
	}

	return &map_result
}

func (this *FinanceApi) genSignString(mapBody map[string]interface{}, pwd string) (string, error) {
	sorted_keys := make([]string, 0)
	for k, _ := range mapBody {
		if k == "access_token" || k == "scope" || k == "sign" {
			continue
		}
		sorted_keys = append(sorted_keys, k)
	}
	sort.Strings(sorted_keys)
	var signStrings string

	index := 0
	for _, k := range sorted_keys {
		value := ghostlib.ToString(mapBody[k])
		signStrings = signStrings + k + "=" + value

		if index < len(sorted_keys)-1 {
			signStrings = signStrings + "&"
		}
		index++
	}

	signStrings = signStrings + "&key=" + pwd

	return signStrings, nil
}

func (this *FinanceApi) makeSign(will_sign_str string) string {
	MY_PRIKEY, _ := ioutil.ReadFile("/data0/xxkey/private_key.pem")
	keyobj, _ := pem.Decode(MY_PRIKEY)
	if keyobj == nil {
		panic("载入RSA私钥错误")
	}

	private, err := x509.ParsePKCS8PrivateKey(keyobj.Bytes)
	if err != nil {
		panic("处理证书失败")
	}

	h := crypto.Hash.New(crypto.SHA1)
	h.Write([]byte(will_sign_str))
	hashed := h.Sum(nil)

	// 进行rsa加密签名
	signedData, err := rsa.SignPKCS1v15(rand.Reader, private.(*rsa.PrivateKey), crypto.SHA1, hashed)
	if err != nil {
		fmt.Println("Error from signing: %s\n", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(signedData)
}

func (this *FinanceApi) verifyArtxunSign(sign, src string) (bool, error) {
	ARTXUN_PUBKEY := []byte(`
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwAbf7SeuIbW8JfyJXjqC
xBZadZP/y90VkYCYMX6jWiyA9me/ugVeUACdkMviZp/x8qbagEgOTl9g6dkkneQT
.......
-----END PUBLIC KEY-----
`)

	pubkeyobj, _ := pem.Decode([]byte(ARTXUN_PUBKEY))
	pubobj, err := x509.ParsePKIXPublicKey(pubkeyobj.Bytes)
	if err != nil {
		fmt.Printf("Failed to parse RSA public key: %s\n", err)
		return false, err
	}
	rsaPub, _ := pubobj.(*rsa.PublicKey)

	h := crypto.Hash.New(crypto.SHA1)
	h.Write([]byte(src))
	digest := h.Sum(nil)

	data, _ := base64.StdEncoding.DecodeString(string(sign))
	//	hexSig := hex.EncodeToString(data)
	//	fmt.Printf("base decoder: %v, %v\n", string(sign), hexSig)

	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA1, digest, data)
	if err != nil {
		fmt.Println("Verify sig error, reason: ", err)
		return false, err
	}

	return true, nil
}

func (this *FinanceApi) GetUserBalance(artxun_openid string) *map[string]interface{} {
	post_arg := map[string]interface{}{
		"scope":         "balance",
		"artxun_openid": artxun_openid,
	}

	sign_str, _ := this.genSignString(post_arg, pwd_charge)
	post_arg["sign"] = this.makeSign(sign_str)
	return this.doFinanceApi(post_arg)
}

//----------------------------------------------------------
//func main() {
//	oid := "579618732147a2ecacb3244a85eb18b5c49240d8"
//	//	userapi := NewUserApi()
//	//	x := userapi.getUserInfo(oid)
//	//	fmt.Printf("%v\n", x)

//	//	fmt.Println(base64.StdEncoding.EncodeToString([]byte("mysql:host=192.168.3.127")))

//	fapi := NewFinanceApi()
//	r := fapi.getUserBalance(oid)
//	fmt.Printf("%v\n", r)

//	//	fmt.Printf("\n\n")
//	//	s := fapi.makeSign("mysql:host=192.168.3.127")
//	//	fmt.Println(s)

//	return

//}
