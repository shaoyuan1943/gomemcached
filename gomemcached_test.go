package gomemcached

import (
	"sync"
	"testing"
	"time"
)

var (
	Once   sync.Once
	Client *MemcachedClient
)

func Instance() *MemcachedClient {
	Once.Do(func() {
		Client = NewMemcachedClient([]string{"192.168.2.169:11211"}, 5)
	})

	return Client
}

type Person struct {
	Name   string
	Age    uint8
	Family []string
}

func setValue(t *testing.T, key string, value interface{}) {
	_, err := Instance().Set(key, value, 0)
	if err != nil {
		t.Errorf("Set err: %v", err)
	}
}

func getValue(t *testing.T, key string, value interface{}) {
	_, err := Instance().Get(key, value)
	if err != nil {
		t.Errorf("Get err: %v", err)
		return
	}

	t.Logf("Get: %v", value)
}

func TestSetAndGetTypeValue(t *testing.T) {
	p := &Person{
		Name:   "lennon",
		Age:    29,
		Family: []string{"Father", "Mother", "Brother"},
	}

	setValue(t, "TestGet_int", -1024)
	var value int
	_, err := Instance().Get("TestGet_int", &value)
	if err != nil {
		t.Errorf("TestGet_int err: %v", err)
		return
	}
	t.Logf("TestGet_int: %v", value)

	setValue(t, "TestGet_uint", 1024)
	var value1 uint
	_, err = Instance().Get("TestGet_uint", &value1)
	if err != nil {
		t.Errorf("TestGet_uint err: %v", err)
		return
	}
	t.Logf("TestGet_uint: %v", value1)

	setValue(t, "TestGet_string", "HellowWorld")
	var value2 string
	_, err = Instance().Get("TestGet_string", &value2)
	if err != nil {
		t.Errorf("TestGet_string err: %v", err)
		return
	}
	t.Logf("TestGet_string: %v", value2)

	setValue(t, "TestGet_uint64", 2817283041904798650)
	var value3 uint64
	_, err = Instance().Get("TestGet_uint64", &value3)
	if err != nil {
		t.Errorf("TestGet_uint64 err: %v", err)
		return
	}
	t.Logf("TestGet_uint64: %v", value3)

	setValue(t, "TestGet_int64", -2817283041904798650)
	var value4 int64
	_, err = Instance().Get("TestGet_int64", &value4)
	if err != nil {
		t.Errorf("TestGet_int64 err: %v", err)
		return
	}
	t.Logf("TestGet_int64: %v", value4)

	setValue(t, "TestGet_bool", false)
	var value5 bool
	_, err = Instance().Get("TestGet_bool", &value5)
	if err != nil {
		t.Errorf("TestGet_bool err: %v", err)
		return
	}
	t.Logf("TestGet_bool: %v", value5)

	var f32 float32 = 6.89
	setValue(t, "TestGet_float32", f32)
	var value6 float32
	_, err = Instance().Get("TestGet_float32", &value6)
	if err != nil {
		t.Errorf("TestGet_float32 err: %v", err)
		return
	}
	t.Logf("TestGet_float32: %v", value6)

	var f64 float64 = 281728.3041904798650
	setValue(t, "TestGet_float64", f64)
	var value7 float64
	_, err = Instance().Get("TestGet_float64", &value7)
	if err != nil {
		t.Errorf("TestGet_float64 err: %v", err)
		return
	}
	t.Logf("TestGet_float64: %v", value7)

	setValue(t, "TestGet_struct", p)
	var value8 Person
	_, err = Instance().Get("TestGet_struct", &value8)
	if err != nil {
		t.Errorf("TestGet_struct err: %v", err)
		return
	}
	t.Logf("TestGet_struct: %v", value8)
}

func TestAtomic(t *testing.T) {
	value, _, err := Instance().Increment("TestAtomic_incr", 10000, 0, 0)
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	value, _, err = Instance().Increment("TestAtomic_incr", 20000, 0, 0)
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	value, _, err = Instance().Increment("TestAtomic_incr", 30000, 0, 0)
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	value, _, err = Instance().Increment("TestAtomic_incr", 40000, 0, 0)
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	val2, err := Instance().TouchAtomicValue("TestAtomic_incr")
	if err != nil {
		t.Errorf("TestAtomic_incr_touch err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr_touch: %v", val2)

}

func TestSetExpiration(t *testing.T) {
	_, err := Instance().Set("TestSetExpiration", "HelloWorld", 10)
	if err != nil {
		t.Errorf("TestSetExpiration err: %v", err)
		return
	}

	<-time.After(time.Second * 2)

	var value string
	_, err = Instance().Get("TestSetExpiration", &value)
	if err != nil {
		t.Errorf("TestSetExpiration err: %v", err)
		return
	}

	t.Logf("TestSetExpiration: %v", value)
}
