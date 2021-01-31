# agingMap
<a href="https://pkg.go.dev/github.com/520MianXiangDuiXiang520/agingMap"> <img src="https://img.shields.io/badge/godo-creferenceref-blue" /></a>

基于 sync.Map 的一个带过期时间的 Map



```go
go get -u github.com/520MianXiangDuiXiang520/agingMap
```



## 使用

```go
packege main

import (
    "github.com/520MianXiangDuiXiang520/agingMap"
    "time"
    "fmt"
)

func main() {
    // 每秒遍历 50%，过期删除
    // am := NewAgingMap()
    
    // 惰性删除
    // am := NewWithLazyDelete()
    
    am := NewBaseAgingMap(time.Second * 5, 0.7)
    keyChan := make(chan int64, 10)
    go func() {
        for {
            key := time.Now().UnixNano()
            keyChan <- key
            am.Store(key, 1, time.Second) 
        }
    }()
    for {
        key := <- keyChan
        val, ok := am.Load(key)
        if ok {
            fmt.Println(val)
        }
    }
}
```

