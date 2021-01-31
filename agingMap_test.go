package agingMap

import (
    `fmt`
    `math/rand`
    `time`
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
            for  {
                key := fmt.Sprintf("%d: %d", i, time.Now().UnixNano())
                ch <- key
                am.Store(key, i, time.Second)
                time.Sleep(time.Duration(rand.Int63n(2000)) * time.Millisecond)
            }
        }(i)
    }
    for i := 0; i < 10; i++ {
        go func(i int) {
            for  {
                key := <- ch
                val, ok := am.Load(key)
                fmt.Println(val, ok)
            }
        }(i)
    }
    for  {
        key := <- ch
        val, ok := am.Load(key)
        fmt.Println(val, ok)
    }
    
}

