package factory

import (
	"bookstore/store"
	"fmt"
	"sync"
)

/*生产 Store的工厂，主要提供了两个方法：
  1、注册store到工厂
  2、根据 name 从工厂获取 Store*/

var (
	//1、为何选择读写锁？读多写少？
	providersMu sync.RWMutex
	//使用map 根据 name 封装不同类型 的 Store
	//采用了一个map 类型数据，对工厂可以“生产”的、满足 Store 接口的实例类型进行管理。
	providers = make(map[string]store.Store)
)

// 提供了 Register 函数，让各个实现 Store 接口的类型可以把自己“注册”到工厂中来。
func Register(name string, p store.Store) {
	//注册用写锁
	providersMu.Lock()
	defer providersMu.Unlock()
	//2、此处为啥不用error，而采用panic，二者一般如何选择？？
	//可能是 由于 需要在init函数中调用，而 init函数没有参数和返回值，不好处理 error
	if p == nil {
		panic("store: Register provider is nil")
	}

	if _, dup := providers[name]; dup {
		panic("store: Register called twice for provider " + name)
	}
	providers[name] = p
}

//一旦注册成功，factory 包就可以“生产”出这种满足 Store 接口的类型实例。而依赖 Store
//接口的使用方，只需要调用 factory 包的 New 函数，再传入期望使用的图书存储实现的名
//称，就可以得到对应的类型实例了。

func New(providerName string) (store.Store, error) {
	//实际上是获取 store，使用读锁
	providersMu.RLock()
	p, ok := providers[providerName]
	providersMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("store: unknown provider %s", providerName)
	}

	return p, nil
}
