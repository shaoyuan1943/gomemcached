package gomemcached

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	Once   sync.Once
	client Client
)

func Instance() Client {
	Once.Do(func() {
		client = NewMemcachedClient([]string{"192.168.2.169:11211"}, 5)
	})

	return client
}

type Person struct {
	Name   string
	Age    uint8
	Family []string
}

func setValue(t *testing.T, key string, value interface{}) {
	_, err := Instance().Set(&KeyArgs{Key: key, Value: value, Expiration: 0, CAS: 0})
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
	value, _, err := Instance().Increment(&KeyArgs{Key: "TestAtomic", Delta: 10000})
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	value, _, err = Instance().Increment(&KeyArgs{Key: "TestAtomic", Delta: 20000})
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	value, _, err = Instance().Increment(&KeyArgs{Key: "TestAtomic", Delta: 30000})
	if err != nil {
		t.Errorf("TestAtomic_incr err: %v", err)
		return
	}
	t.Logf("TestAtomic_incr: %v", value)

	value, _, err = Instance().Increment(&KeyArgs{Key: "TestAtomic", Delta: 40000})
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
	_, err := Instance().Set(&KeyArgs{Key: "TestSetExpiration", Value: "HelloWorld", Expiration: 10})
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

func TestAppend(t *testing.T) {
	_, err := Instance().SetRawData(&KeyArgs{Key: "TestAppend", Value: []byte("HelloWorld")})
	if err != nil {
		t.Errorf("TestAppend err: %v", err)
		return
	}

	_, err = Instance().Append(&KeyArgs{Key: "TestAppend", Value: []byte("Iamprogrammer")})
	if err != nil {
		t.Errorf("TestAppend err: %v", err)
		return
	}

	var value1 []byte
	_, err = Instance().Get("TestAppend", &value1)
	if err != nil {
		t.Errorf("TestAppend err: %v", err)
		return
	}

	t.Logf("TestAppend value1: %v", string(value1))

	_, err = Instance().Prepend(&KeyArgs{Key: "TestAppend", Value: []byte("NiceTooMeetYou")})
	if err != nil {
		t.Errorf("TestAppend err: %v", err)
		return
	}

	var value2 []byte
	_, err = Instance().Get("TestAppend", &value2)
	if err != nil {
		t.Errorf("TestAppend err: %v", err)
		return
	}

	t.Logf("TestAppend value2: %v", string(value2))
}

func TestCAS(t *testing.T) {
	cas, err := Instance().Add(&KeyArgs{Key: "TestCAS3", Value: "HelloWorld"})
	if err != nil {
		t.Errorf("Set err-->1: %v", err)
		return
	}
	t.Logf("cas-->1: %v", cas)

	cas, err = Instance().Add(&KeyArgs{Key: "TestCAS3", Value: "NiceTooMeetYou", CAS: cas})
	if err != nil {
		t.Errorf("Set err-->2: %v", err)
		return
	}
	t.Logf("cas-->2: %v", cas)

	var value string
	cas, err = Instance().Get("TestCAS3", &value)
	t.Logf("Get: %v, %v", cas, value)

	cas, err = Instance().Add(&KeyArgs{Key: "TestCAS3", Value: "Iamironman", CAS: cas + 1})
	if err != nil {
		t.Errorf("Set err-->3: %v", err)
		return
	}
	t.Logf("cas-->3: %v", cas)
}

func TestFlush(t *testing.T) {
	_, err := Instance().Set(&KeyArgs{Key: "TestFlush1", Value: "yuriyiuq"})
	if err != nil {
		t.Fatalf("TestFlush1 err: %v", err)
		return
	}

	_, err = Instance().Set(&KeyArgs{Key: "TestFlush2", Value: "fshjkfsjk"})
	if err != nil {
		t.Fatalf("TestFlush2 err: %v", err)
		return
	}

	_, err = Instance().Set(&KeyArgs{Key: "TestFlush3", Value: "uioufsjkfjs"})
	if err != nil {
		t.Fatalf("TestFlush3 err: %v", err)
		return
	}

	err = Instance().Flush(&KeyArgs{})
	if err != nil {
		t.Fatalf("Flush err: %v", err)
		return
	}

	var value string
	_, err = Instance().Get("TestFlush3", &value)
	t.Logf("Get TestFlush3: %v, %v", value, err)

	_, err = Instance().Set(&KeyArgs{Key: "TestFlush5", Value: "gdsgsdfgsd"})
	if err != nil {
		t.Fatalf("TestFlush2 err: %v", err)
		return
	}

	_, err = Instance().Set(&KeyArgs{Key: "TestFlush6", Value: "gsdfgvxcvadg"})
	if err != nil {
		t.Fatalf("TestFlush3 err: %v", err)
		return
	}

	err = Instance().Flush(&KeyArgs{Expiration: 10})
	if err != nil {
		t.Fatalf("Flush err: %v", err)
		return
	}

	var val string
	_, err = Instance().Get("TestFlush6", &val)
	t.Logf("Get TestFlush6: %v, %v", val, err)

	<-time.After(time.Second * 12)
	val = ""
	_, err = Instance().Get("TestFlush6", &val)
	t.Logf("Get TestFlush6: %v, %v", val, err)
}

var rander *rand.Rand

func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	for i := 0; i < l; i++ {
		result = append(result, bytes[rander.Intn(len(bytes))])
	}
	return string(result)
}

func BenchmarkMemcachedClient(b *testing.B) {
	b.StopTimer()

	rander = rand.New(rand.NewSource(time.Now().UnixNano()))
	benchTimes := b.N
	var randomKey []string
	var randomValue []string
	for i := 0; i < benchTimes; i++ {
		randomKey = append(randomKey, GetRandomString(10))
		randomValue = append(randomValue, GetRandomString(8))
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := Instance().Set(&KeyArgs{Key: randomKey[i], Value: randomValue[i]})
		if err != nil {
			b.Fatalf("Set err: %v", err)
			return
		}
	}
}
