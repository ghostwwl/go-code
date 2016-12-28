/**
 * User: ghostwwl
 * Date: 16-12-4
 * Time: 上午11:17
 */


package main

import (
	//"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	//"net"
	//"strings"
	"strings"
	"os"
)

func main() {
	//runHttpService()
	runHttpService2()
}



func PathExists(fpath string) bool {
	_, err := os.Stat(fpath)
	if nil == err {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	return false
}


func runHttpService2() {
	mux := http.NewServeMux()

	mux.HandleFunc("/h", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Go net/http Test!!!"))
	})

	mux.HandleFunc("/bye", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "byebye")
	})

	mux.HandleFunc("/hello", handle_hello)
	mux.HandleFunc("/index", handle_index)
	mux.HandleFunc("/uptime", handle_uptime)
	mux.HandleFunc("/tobaidu", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://www.baidu.com", http.StatusFound)
	})



	STATIC_PATH_MAP := make(map[string]string)
	STATIC_PATH_MAP["/src"] = "./test"
	STATIC_PATH_MAP["/baidu"] = "/data/ghostwwl/html-百度"


	// 其它情况的handle呢
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 静态文件服务
		path_array_list := strings.Split(r.URL.Path, "/")
		alias_path := "/" + path_array_list[1]
		localdir, has_static_map := STATIC_PATH_MAP[alias_path];
		if has_static_map {
			local_file_path := localdir + r.URL.Path[len(alias_path):]
			log.Printf("Get File:%s", local_file_path)
			if PathExists(local_file_path) {
				http.ServeFile(w, r, local_file_path)
			} else {
				http.NotFound(w, r)
			}
			return
		}
		// end 静态文件


		// 获取 cookie
		mycv, xer := r.Cookie("testcook")
		if nil != xer{
			fmt.Fprintf(w, "%v", xer)
		} else {
			fmt.Fprintf(w, "%v", mycv.Value)
		}

		// 跳转
		//w.Header().Set("Location", "/uptime")
		//w.WriteHeader(302)

		// use cookie
		cookie := http.Cookie{
			Name: "testcook",
			Value: "123",
			Path: "/",
		}
		http.SetCookie(w, &cookie)


		log.Println(r.RequestURI)

		//http.Error(w, "我靠 404 了", 404)


		fmt.Fprintf(w, "我靠 404 了\n")
		fmt.Fprintf(w, "Thanks for the %s query: %s", r.Method, r.URL.RawQuery)
		fmt.Fprintf(w, "\npath: ", r.Method, r.URL.Path)
		fmt.Fprintf(w, "\nYou IP: %s", r.RemoteAddr)
		fmt.Fprintf(w, "\nUserAgent: %s", r.UserAgent())

		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "\nerror: %v", err)
			//return
		}
		fmt.Fprintf(w, "\nparam: %v", r.Form.Get("a"))
		fmt.Fprintf(w, "\nparam: %v", r.FormValue("b"))
		fmt.Fprintf(w, "\nGET: %v", r.URL.Query())
		fmt.Fprintf(w, "\nGET: %v", r.Form)
		fmt.Fprintf(w, "\nPOST: %v", r.PostForm)
		err = r.ParseMultipartForm(16 << 10) // 16 mb
		if nil != err {
			fmt.Fprintf(w, "\nParseMultipartForm Error: %v", err)
		} else {
			fmt.Fprintf(w, "\nMutilPOST: %v", r.MultipartForm.Value)
		}

		// r.URL.Query() --> 获取到的是 GET 的MAP
		// r.ParseForm() 之后
		// r.Form 获取到的是 GET 和 url-encode 方式POST的 MAP
		// r.PostForm 获取到 url-encode post的 参数

		// r.ParseMultipartForm(16 << 10) // 16 mb -- 如果connected不是 multipart/form-data 会报错
		// r.MultipartForm.Value 获取到的是 multipart/form-data 上传的 MAP 这个时候 r.PostForm 是空的

	})


	//http.Request

	http.ListenAndServe(":8001", mux)
}

//---------------------------------------------------------------------
func runHttpService() {
	//	http.HandleFunc("/index", handle_index)
	//	http.HandleFunc("/hello", handle_hello)
	//	http.ListenAndServe(":8001", nil)

	http.ListenAndServe(":8001", &http_handle{})
}

type http_handle struct{}

func (*http_handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.String() //获得访问的路径
	//io.WriteString(w, path)

	switch r.URL.String() {
	case "/hello":
		handle_hello(w, r)
		break
	case "/index":
		handle_index(w, r)
		break
	default:
		w.Write([]byte(path))
	}

}

//---------------------------------------------------------------------

func handle_index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("你吃了么?\n"))
	w.Write([]byte("我还没吃?"))
}

func handle_hello(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Hello"))
}

func handle_uptime(w http.ResponseWriter, req *http.Request) {
	cmd := exec.Command("uptime")
	//cmd2 := exec.Command("last")

	bytes, err := cmd.Output()
	if err != nil {
		//fmt.Println("cmd.Output: ", err)
		fmt.Fprintf(w, "出错了: %s", err)
	} else {
		fmt.Fprintf(w, "服务器uptime为: %s\n", string(bytes))
	}

	//lastr, err := cmd2.Output()
	//if err != nil {
	//	//fmt.Println("cmd.Output: ", err)
	//	fmt.Fprintf(w, "出错了: %s", err)
	//} else {
	//	fmt.Fprintf(w, "服务器last为: %s", string(lastr))
	//}
}
