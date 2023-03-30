package server

import (
	"bookstore/server/middleware"
	"bookstore/store"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

//HTTP 服务模块的职责是对外提供 HTTP API 服务，处理来自客户端的各种请求，并通过
//Store 接口实例执行针对图书数据的相关操作。这里，我们抽象处理一个 server 包，这个包
//中定义了一个 BookStoreServer

// 这个类型实质上就是一个标准库的 http.Server，并且组合了来自 store.Store 接口的能力。
// BookStoreServer 由 store.Store 和 http.Server组合而成，拥有了二者的能力
type BookStoreServer struct {
	s   store.Store
	srv *http.Server
}

//server 包提供了 NewBookStoreServer 函数，用来创建一个 BookStoreServer类型实例
//我们看到函数 NewBookStoreServer 接受两个参数，一个是 HTTP 服务监听的服务地址，另
//外一个是实现了 store.Store 接口的类型实例。这种函数原型的设计是 Go 语言的一种惯用设
//计方法，也就是接受一个接口类型参数，返回一个具体类型。返回的具体类型组合了传入的接
//口类型的能力。

func NewBookStoreServer(addr string, s store.Store) *BookStoreServer {
	srv := &BookStoreServer{
		s: s,
		srv: &http.Server{
			//host:port
			Addr: addr,
		},
	}

	//由于这个服务请求 URI 的模式字符串比较复杂，标准库 http 包内置的 URI 路径模式匹配器
	//（ServeMux，也称为路由管理器）不能满足我们的需求，因此在这里，我们需要借助一个第
	//三方包 github.com/gorilla/mux 来实现我们的需求。
	router := mux.NewRouter()
	router.HandleFunc("/book", srv.createBookHandler).Methods("POST")
	router.HandleFunc("/book/{id}", srv.updateBookHandler).Methods("POST")
	router.HandleFunc("/book/{id}", srv.getBookHandler).Methods("GET")
	router.HandleFunc("/book", srv.getAllBooksHandler).Methods("GET")
	router.HandleFunc("/book/{id}", srv.delBookHandler).Methods("DELETE")

	//我们在 router 的外围包裹了两层 middleware。什么是 middleware
	//呢？对于我们的上下文来说，这些 middleware 就是一些通用的 http 处理函数。

	// mux.Router 实现了 http.Handler 接口，因此它与标准的 http.ServeMux 兼容
	//对 router 增加了 记录日志 和 验证的功能
	srv.srv.Handler = middleware.Logging(middleware.Validating(router))
	return srv
}

// 这个函数把 BookStoreServer 内部的 http.Server 的运行，放置到一个单独的轻
// 量级线程 Goroutine 中。这是因为，http.Server.ListenAndServe 会阻塞代码的继续运行，
// 如果不把它放在单独的 Goroutine 中，后面的代码将无法得到执行。
func (bs *BookStoreServer) ListenAndServe() (<-chan error, error) {
	var err error
	//为了检测到 http.Server.ListenAndServe 的运行状态，我们再通过一个 channel
	//在新创建的 Goroutine 与主 Goroutine 之间建立的通信渠道。通过这个渠
	//道，这样我们能及时得到 http server 的运行状态。
	errChan := make(chan error)
	go func() {
		err = bs.srv.ListenAndServe()
		errChan <- err
	}()

	select {
	//启动 go routine之后 马上检查，是否有err，有err，那么就是 启动 err，直接返回err
	case err = <-errChan:
		return nil, err
	//如果没有 err，那么就返回 errChan，后续errChan 中有err，就属于运行时 的 err了
	case <-time.After(time.Second):
		return errChan, nil
	}
}

func (bs *BookStoreServer) Shutdown(ctx context.Context) error {
	return bs.srv.Shutdown(ctx)
}

func (bs *BookStoreServer) createBookHandler(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	var book store.Book
	//create book 请求体 解码成 store.Book
	if err := dec.Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 调用 Store 的 create 方法，存储 store.Book
	if err := bs.s.Create(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("create ok"))
}

func (bs *BookStoreServer) updateBookHandler(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "no id found in request", http.StatusBadRequest)
		return
	}

	dec := json.NewDecoder(req.Body)
	var book store.Book
	if err := dec.Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	book.Id = id
	if err := bs.s.Update(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte("update ok"))
}

func (bs *BookStoreServer) getBookHandler(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "no id found in request", http.StatusBadRequest)
		return
	}

	book, err := bs.s.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response(w, book)
}

func (bs *BookStoreServer) getAllBooksHandler(w http.ResponseWriter, req *http.Request) {
	books, err := bs.s.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response(w, books)
}

func (bs *BookStoreServer) delBookHandler(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "no id found in request", http.StatusBadRequest)
		return
	}

	err := bs.s.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte("delete ok"))
}

func response(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")

	//使用 Encode 方式编码
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//使用 Marshal 方式编码
	//data, err := json.Marshal(v)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Write(data)

}
