package main

import (
	_ "bookstore/internal/store"
	"bookstore/server"
	"bookstore/store/factory"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	s, err := factory.New("mem")
	//1、为啥直接 panic 了，记录日志也没啥用？？？
	if err != nil {
		panic(err)
	}

	srv := server.NewBookStoreServer("192.168.30.58:8080", s)

	errChan, err := srv.ListenAndServe()
	//如果启动后马上检出 err，那么可以认为是 web 启动错误
	if err != nil {
		log.Println("web server start failed:", err)
		return
	}
	log.Println("web server start ok")

	//通过 signal 包的 Notify 捕获了 SIGINT、SIGTERM 这两个
	//系统信号。这样，当这两个信号中的任何一个触发时，我们的 http 服务实例都有机会在退出
	//前做一些清理工作。
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	select {
	// web server 启动后，errChan有错误，可以看成是 web 运行时 错误
	case err = <-errChan:
		log.Println("web server run failed:", err)
		return
	//当 c 中 能取出内容时候，证明发生了 syscall.SIGINT 或 syscall.SIGTERM
	case <-c:
		log.Println("bookstore program is exiting...")
		ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
		//尽管ctx会过期，但在任何情况下调用它的cancel函数都是很好的实践，（如果 srv.Shutdown(ctx)的 业务 提前完成了，就可以立即结束了，否则至少需要等待 time.Second时间）
		//如果不这样做，可能会使上下文及其父类存活的时间超过必要的时间
		defer cf()
		//srv.Shutdown(ctx)，可能会返回错误
		//如果提供的上下文在关闭完成之前过期，则 Shutdown 返回上下文的错误
		//否则它返回关闭服务器的底层侦听器返回的任何错误。
		err = srv.Shutdown(ctx)
	}

	if err != nil {
		log.Println("bookstore program exit error:", err)
		return
	}
	log.Println("bookstore program exit ok")

}
