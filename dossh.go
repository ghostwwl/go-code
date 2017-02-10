/*****************************************
 * FileName  : dossh.go
 * Author    : ghostwwl
 * Date      : 2016.12.07
 * Note      : 在远程 ssh 上执行一个命令并返回结果
 * History   :
 *****************************************/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func getClient(user, addr, passwd string) *ssh.Client {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(passwd)},
	}
	client, err := ssh.Dial("tcp", addr, config)
	if nil != err {
		fmt.Printf("connect failed.(%v)\n", err)
		os.Exit(-1)
	}
	return client
}

func doCmd(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	session.Stdout = &buf
	err = session.Run(cmd)
	if err != nil {
		return "", err
	}
	defer session.Close()
	return string(buf.Bytes()), nil
}

func main() {
	user := flag.String("user", "root", "linux system user default root")
	host := flag.String("host", "", "the linux host and port like 'ip:port' ")
	port := flag.Int("port", 22, "the client port default 22 ")
	pwd := flag.String("pwd", "", "the system password")
	cmd := flag.String("c", "", "the cmd want to run on remoter")

	flag.Parse()
	if "" == *pwd || "" == *cmd || "" == *host {
		flag.Usage()
		return
	}
	dst_host := fmt.Sprintf("%s:%v", *host, *port)
	do_result, err := doCmd(getClient(*user, dst_host, *pwd), *cmd)
	if nil != err {
		fmt.Printf("run cmd error.(%v)\n", err)
		os.Exit(-2)
	}
	fmt.Println(do_result)
}
