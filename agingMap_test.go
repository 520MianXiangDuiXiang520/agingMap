package agingMap

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func ExampleAgingMap_Delete() {
	am := NewAgingMap()
	am.Store("key", "value", time.Second)
	am.Delete("key")
}

func ExampleAgingMap_Store() {
	am := NewAgingMap()
	am.Store("key", "value", time.Second)
}

func ExampleAgingMap_Load() {
	am := NewAgingMap()
	ch := make(chan string, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				key := fmt.Sprintf("%d: %d", i, time.Now().UnixNano())
				ch <- key
				am.Store(key, i, time.Second)
				time.Sleep(time.Duration(rand.Int63n(2000)) * time.Millisecond)
			}
		}(i)
	}
	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				key := <-ch
				val, ok := am.Load(key)
				fmt.Println(val, ok)
			}
		}(i)
	}
	for {
		key := <-ch
		val, ok := am.Load(key)
		fmt.Println(val, ok)
	}
}

func TestAgingMap(t *testing.T) {
	aMap := NewWithLazyDelete()
	aMap.Store("key", "val", time.Second)
	time.Sleep(time.Second)
	v, ok := aMap.Load("key")
	if ok || v != nil {
		t.Error("get expired data")
	}
}

func TestAgingMap_AutoDelete(t *testing.T) {
	aMap := NewBaseAgingMap(time.Second, 1)
	for i := 0; i < 7; i++ {
		aMap.Store(i, "val", time.Second)
	}
	time.Sleep(time.Second * 2)
	for i := 0; i < 7; i++ {
		v, ok := aMap._map.Load(i)
		if ok || v != nil {
			t.Error("get expired data")
		}
	}
}

func TestAgingMap_LoadOrStore(t *testing.T) {
	aMap := NewBaseAgingMap(time.Second, 1)
	_, _, stored := aMap.LoadOrStore("key", 1, time.Second)
	if !stored {
		t.Errorf("第一次未存储")
	}
	v, _, stored := aMap.LoadOrStore("key", 1, time.Second)
	if v != 1 || stored {
		t.Errorf("第二次存储")
	}
	time.Sleep(time.Second)
	_, _, stored = aMap.LoadOrStore("key", 1, time.Second)
	if !stored {
		t.Errorf("第一次未存储")
	}
}

func TestAgingMap_LoadOrStore_concurrent(t *testing.T) {
	aMap := NewBaseAgingMap(time.Second, 1)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		var v1, v2 interface{}
		var s1, s2 bool
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			v1, _, s1 = aMap.LoadOrStore(i, fmt.Sprintf("F%d", i), time.Second)
		}(i)
		go func(i int) {
			defer wg.Done()
			v2, _, s2 = aMap.LoadOrStore(i, fmt.Sprintf("S%d", i), time.Second)
		}(i)
		wg.Wait()
		if v1 != v2 {
			t.Errorf("两次值一样， V1 = %v, V2 = %v", v1, v2)
		}
		if s1 && s2 {
			t.Errorf("true true")
		}
		if !(s1 || s2) {
			t.Errorf("false false")
		}
	}
}

func TestAgingMap_Store(t *testing.T) {
	aMap := NewBaseAgingMap(time.Minute, 0.5)
	go func() {
		for i := 0; i < 7; i++ {
			aMap.Store(i, "val", time.Second*10)
			fmt.Println("Store: ", i)
			time.Sleep(10 * time.Second)
		}
	}()
	time.Sleep(45 * time.Second)
	aMap.Range(func(k, v interface{}) bool {
		fmt.Println(k, v)
		return true
	})

	fmt.Println("------")
	time.Sleep(20 * time.Second)
	aMap.Range(func(k, v interface{}) bool {
		fmt.Println(k, v)
		return true
	})

}

func TestAgingMap_LoadWithDeadline(t *testing.T) {
	am := NewAgingMap()
	am.Store(1, 2, time.Minute)
	for i := 0; i < 70; i++ {
		fmt.Println(am.LoadWithDeadline(1))
		time.Sleep(time.Second * 10)
	}
}
