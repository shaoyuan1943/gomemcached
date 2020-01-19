package main

import (
	"fmt"
	"gomemcached"
)

func main() {
	m := gomemcached.NewMemcachedClient([]string{"10.11.133.161:11211",
		"10.11.133.161:11212", "10.11.133.161:11213", "10.11.133.161:11214"}, 1)
	err := m.Set("Name", "chencheng", 0, 0)
	if err != nil {
		fmt.Printf("Set err->1: %v\n", err)
		return
	}

	err = m.Set("Name", "xiaoyanni", 0, 500)
	if err != nil {
		fmt.Printf("Set err->2: %v\n", err)
		return
	}

	err = m.Set("Name", "zhoujielun", 0, 0)
	if err != nil {
		fmt.Printf("Set err->3: %v\n", err)
		return
	}

	var name string
	cas, err := m.Get("Name", &name)
	if err != nil {
		fmt.Printf("Get err: %v\n", err)
		return
	}
	fmt.Printf("Get: %v, cas: %v\n", name, cas)

}
