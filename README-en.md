[English](./README.md)/[中文](./README-zh.md)

### gomemcached
gomemcached is a light-weight and high performance memcache client for Go.   
* Binary protocol  
* Support sync and async mode  
* Sample interface

### Install
```go get github.com/shaoyuan1943/gomemcached```

### How to use
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
gomemcached use msgpack to serialize data. For `MemcachedClient`, every operation will be serialized by msgpack into a complete data package. In theory, it is not possible to add or subtract some data to this data package. When you need to increase or decrease part of the data, use the `* RawData` function. Such functions do not use msgpack to serialize the data.  

#### 参数
`expiration` Expiration time in seconds  
`cas` If the Data Version Check (CAS) is nonzero, the requested operation MUST only succeed if the item exists and has a CAS value identical to the provided value.  

#### 函数
`func (m *MemcachedClient) AddServer(addr string, maxConnPerServer uint32) error`    
Add a memcached server, the error is nil when the operation is successful.    

`func (m *MemcachedClient) SetServerErrorCallback(call ServerErrorCallback)`    
Set callback when memcached server failed, the callback's parameter is server address.     

`func (m *MemcachedClient) Exit()`    
Exit client by manual control, in theory, that client will not be available after this function is called.    

`func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error)`    
Get the value of key, `value` is a pointer to a value variable. Return value is the CAS corresponding to the key, and the error is nil when the operation is successful.    

`func (m *MemcachedClient) Set(key string, value interface{}, expiration uint32, cas uint64) (uint64, error)`   Set the value of key. Return value is the CAS corresponding to the key, and the error is nil when operation is successful.    

`func (m *MemcachedClient) SetRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error)`    
Same as `Set`, when increase or decrease part of the data, must use this function. Return value is CAS, the error is nil when operation is successful, the function does not serialize data.    

`func (m *MemcachedClient) Add(key string, value interface{}, expiration uint32, cas uint64) (uint64, error)` Add the value of key, return value is CAS, the error is nil when operation is successful.   

`func (m *MemcachedClient) AddRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error)`    
Same as `Add`, when increase or decrease part of the data, must use this function.   

`func (m *MemcachedClient) Replace(key string, value interface{}, expiration uint32, cas uint64) (uint64, error)`   
Replace the value of key, return value is CAS, the error is nil when operation is successful.    

`func (m *MemcachedClient) ReplaceRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error)`    
Same as `Replace`, when increase or decrease part of the data, must use this function.    

`func (m *MemcachedClient) Append(key string, value []byte, cas uint64) (uint64, error)`    
`func (m *MemcachedClient) Prepend(key string, value []byte, cas uint64) (uint64, error)`     
Appends data to the tail/head of an existing value, return value is the CAS, and the error is nil when the operation is successful. This function does not serialize data.    

`func (m *MemcachedClient) Increment(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error)`     
`func (m *MemcachedClient) Decrement(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error)`    
Atomic operation, the delta of the existing value is increased/decreased. If the key does not exist, the operation returns the initial value of the key. The error is nil when the operation is successful.    

`func (m *MemcachedClient) TouchAtomicValue(key string) (uint64, error)`    
Returns the current value of an atom. The error is nil when the operation is successful.    

### 开源协议
MIT License