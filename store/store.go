package store

import "errors"

// 定义了 两个 错误变量，供外部使用
var (
	ErrNotFound = errors.New("not found")
	ErrExist    = errors.New("exist")
)

// 定义图书 以及 图书的 存储相关的接口
// 建立了一个对应图书条目的抽象数据类型 Book，以及针对 Book 存取的接口类型 Store
// 这样，对于想要进行图书数据操作的一方来说，他只需要得到一个满足 Store 接口的实例
// 就可以实现对图书数据的存储操作了，不用再关心图书数据究竟采用了何种存储方式。
// 这就实现了图书存储操作与底层图书数据存储方式的解耦。而且，这种面向接口编程也是 Go 组合设计哲学的一个重要体现。
type Book struct {
	Id      string   `json:"id"`      // 图书ISBN ID
	Name    string   `json:"name"`    // 图书名称
	Authors []string `json:"authors"` // 图书作者
	Press   string   `json:"press"`   // 出版社
}

type Store interface {
	//接口中的 参数和返回值 不需要写名字
	//1、参数 *Book 如果改成 使用 Book可以吗
	Create(*Book) error
	Update(*Book) error
	Get(string) (Book, error)
	GetAll() ([]Book, error)
	Delete(string) error
}
