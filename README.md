[English](./README-en.md)/[中文](./README.md)

### gomemcached
gomemcached定义为轻量级和高性能的memcached Go客户端，特点：  
* 支持二进制协议  
* 支持同步与异步模式（异步模式在v2分支中实现）  
* 易于使用的接口    

### 安装
```go get github.com/shaoyuan1943/gomemcached```

### 如何使用
``` go
func main() {
    m := gomemcached.NewMemcachedClient([]string{"192.168.2.169:11211", []string{"192.168.2.169:112120"}})
    cas, err := m.Set("First", "HelloWorld", 0, 0)
    if err != nil {
        fmt.Printf("Set err: %v\n", err)
        return
    }

    var value string
    cas, err = m.Get("First", &value)
    if err != nil {
        fmt.Printf("Get err: %v\n", err)
        return
    }
}
```

### 接口说明 
gomemcached内部采用msgpack序列化数据，对于`MemcachedClient`而言，每一次操作都会被msgpack序列化为一个完整的数据包，理论上来说，无法对这个数据包增加或减少部分数据。当需要对数据进行增加或减少部分数据的操作时，请使用`*RawData`函数，此类函数不使用msgpack序列化value。  

#### 参数
``` go
type KeyArgs struct {
	Key        string   
	Value      interface{}
	Expiration uint32   // 过期时间，以秒为单位
	CAS        uint64   // 修订号，当cas不为0时：所请求的操作务必仅在key存在且CAS值与提供的值相同时成功
	Delta      uint64   // 原子操作时得步长值
}
```  

#### 接口
**`AddServer(addr string, maxConnPerServer uint32) error`**  
添加一个memcached server，操作成功时返回值为nil 。

**`SetServerErrorCallback(call ServerErrorCallback)`**  
设置某个memcached server失效时的回调函数，该回调函数的参数是失效server的地址。  

**`Exit()`**  
结束该client，理论上来说，此函数调用后该client将无法使用。  

**`Get(key string, value interface{}) (uint64, error)`**  
获取key的值，value是值变量的指针。返回值是key对应的CAS，操作成功时error为nil。  

**`Set(args *KeyArgs) (uint64, error)`**   
设置key的值，返回值是key对应的CAS，操作成功时error为nil。  

**`SetRawData(args *KeyArgs) (uint64, error)`**  
该函数行为与Set一致，当value需要Append或Prepend时，使用该函数完成操作。返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。

**`Add(args *KeyArgs) (uint64, error)`**  
添加某个值，返回值是key对应的CAS，操作成功时error为nil。

**`AddRawData(args *KeyArgs) (uint64, error)`**  
该函数行为与Add一致，当value需要Append或Prepend时，需要使用该函数完成操作。返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。

**`Replace(args *KeyArgs) (uint64, error)`**  
替换key的值。返回值是key对应的CAS，操作成功时error为nil。

**`ReplaceRawData(args *KeyArgs) (uint64, error)`**  
该函数行为与Add一致，当value需要Append或Prepend时，需要使用该函数完成add操作。返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。

**`Append(args *KeyArgs) (uint64, error)`  
`Prepend(args *KeyArgs) (uint64, error)`**  
向一个已存在的值的尾部/首部添加数据，返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。  

**`Increment(args *KeyArgs) (uint64, uint64, error)`**  
**`Decrement(args *KeyArgs) (uint64, uint64, error)`**    
原子操作，已存在的值增加/减少delta，如key不存在则该操作为返回key的初始值，操作成功时error为nil。   

**`TouchAtomicValue(key string) (uint64, error)`**  
返回某个原子的当前值，操作成功时error为nil。

**`Flush(args *KeyArgs) error`**  
清除所有项，当args.Expiration不为0，则表示延迟多少秒后清除。  

### 更多
https://github.com/memcached/memcached/wiki/BinaryProtocolRevamped

### 开源协议
MIT License