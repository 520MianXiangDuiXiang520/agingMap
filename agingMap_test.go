package agingMap

import (
	"fmt"
	"math/rand"
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
