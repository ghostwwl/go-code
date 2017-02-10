package main

/**
Auth: ghostwwl
Email: ghostwwl@gmail.com
**/


import (
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/proxy"
)

func main() {
	// 这里是有密码的
	sio, err := proxy.SOCKS5("tcp", "210.51.190.227:9066", &proxy.Auth{User: "123", Password: "123"}, proxy.Direct)
	// 这里是没有密码的
	sio, err := proxy.SOCKS5("tcp", "210.51.190.227:9066", nil, proxy.Direct)
	if nil != err {
		panic("连不上代理服务器呢")
	}
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}

	httpTransport.Dial = sio.Dial
	resp, err := httpClient.Get("http://www.ip.cn")
	defer resp.Body.Close()

	if nil != err {
		panic(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n", body)
}
