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

### 函数说明 
gomemcached内部采用msgpack序列化数据，对于`MemcachedClient`而言，每一次操作都会被msgpack序列化为一个完整的数据包，理论上来说，无法向这个数据包增加或减少部分数据。当需要对数据进行增加或减少部分数据的操作时，请使用`*RawData`函数，此类函数不使用msgpack序列化value。  

#### 参数
`expiration` 过期时间，以秒为单位。  
`cas` 修订号，当cas不为0时：所请求的操作务必仅在key存在且CAS值与提供的值相同时成功。  

#### 函数
`func (m *MemcachedClient) AddServer(addr string, maxConnPerServer uint32) error`  
添加一个memcached server，操作成功时返回值为nil 。

`func (m *MemcachedClient) SetServerErrorCallback(call ServerErrorCallback)`  
设置某个memcached server失效时的回调函数，该回调函数的参数是失效server的地址。  

`func (m *MemcachedClient) Exit()`  
结束该client，理论上来说，此函数调用后该client将无法使用。  

`func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error)`  
获取key的值，value是值变量的指针。返回值是key对应的CAS，操作成功时error为nil。  

`func (m *MemcachedClient) Set(key string, value interface{}, expiration uint32, cas uint64) (uint64, error)`   
设置key的值，返回值是key对应的CAS，操作成功时error为nil。  

`func (m *MemcachedClient) SetRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error)`  
该函数行为与Set一致，当value需要Append或Prepend时，使用该函数完成操作。返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。

`func (m *MemcachedClient) Add(key string, value interface{}, expiration uint32, cas uint64) (uint64, error)`  
添加某个值，返回值是key对应的CAS，操作成功时error为nil。

`func (m *MemcachedClient) AddRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error)`  
该函数行为与Add一致，当value需要Append或Prepend时，需要使用该函数完成操作。返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。

`func (m *MemcachedClient) Replace(key string, value interface{}, expiration uint32, cas uint64) (uint64, error)`  
替换key的值。返回值是key对应的CAS，操作成功时error为nil。

`func (m *MemcachedClient) ReplaceRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error)`  
该函数行为与Add一致，当value需要Append或Prepend时，需要使用该函数完成add操作。返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。

`func (m *MemcachedClient) Append(key string, value []byte, cas uint64) (uint64, error)`  
`func (m *MemcachedClient) Prepend(key string, value []byte, cas uint64) (uint64, error)`  
向一个已存在的值的尾部/首部添加数据，返回值是key对应的CAS，操作成功时error为nil。此函数不会序列化数据。  

`func (m *MemcachedClient) Increment(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error)`  
`func (m *MemcachedClient) Decrement(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error)`    
原子操作，已存在的值增加/减少delta，如key不存在则该操作为返回key的初始值，操作成功时error为nil。   

`func (m *MemcachedClient) TouchAtomicValue(key string) (uint64, error)`  
返回某个原子的当前值，操作成功时error为nil。

### 开源协议
MIT License