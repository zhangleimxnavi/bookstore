package middleware

import (
	"fmt"
	"log"
	"mime"
	"net/http"
)

//这些 middleware 就是一些通用的 http 处理函数
//Logging 函数主要用来输出每个到达的 HTTP 请求的一些概要信息，而
//Validating 则会对每个 http 请求的头部进行检查，检查 Content-Type 头字段所表示的媒体
//类型是否为 application/json。这些通用的 middleware 函数，会被串联到每个真正的处理
//函数之前，避免我们在每个处理函数中重复实现这些逻辑。

func Logging(next http.Handler) http.Handler {
	//就是 把原来的 http.Handler 增加了 log功能
	//1、一定要用 http.HandlerFunc 吗？是否有其他办法构造 一个新的 http.Handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("recv a %s request from %s", req.Method, req.RemoteAddr)
		next.ServeHTTP(w, req)
	})
}

func Validating(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		method := req.Method
		if method != "GET" && method != "DELETE" {
			contentType := req.Header.Get("Content-Type")
			mediatype, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				fmt.Println("err: ", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if mediatype != "application/json" {
				http.Error(w, "invalid Content-Type", http.StatusUnsupportedMediaType)
				return
			}
			next.ServeHTTP(w, req)
		} else {
			next.ServeHTTP(w, req)
		}
		
	})
}
