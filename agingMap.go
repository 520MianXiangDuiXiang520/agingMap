// Copyright 2009 The Go Junebao. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// AgingMap 提供了一种时效性的并发安全的 Map
package agingMap

import (
	"fmt"
	"github.com/robfig/cron"
	"time"
)

type agingValue struct {
	v          interface{}
	age        time.Duration
	createTime int64
}

type AgingMap struct {
	_map        Map
	task        *cron.Cron
	deleteScale float64
}

func (am *AgingMap) deleteExpiredItems() {
	index := 0
	am._map.Range(func(key, value interface{}) bool {
		index++
		v := value.(agingValue)
		if time.Now().UnixNano()-v.createTime > v.age.Nanoseconds() {
			am.Delete(key)
		}
		return float64(index)/float64(am._map.ReadSize()) < am.deleteScale
	})
}

// NewAgingMap 用来创建一个时效性的 Map.
// NewAgingMap 创建的 Map 每隔 1s 会随机检查 50% 的数据，如果检查到的项目过期
// 他们会被删除，如果想要控制检查速率和范围，请使用 NewBaseAgingMap, 如果不希望
// 主动检查，请使用 NewWithLazyDelete。
func NewAgingMap() *AgingMap {
	return NewBaseAgingMap(time.Second, 0.5)
}

// NewBaseAgingMap 用来创建一个时效性的 Map.
// NewAgingMap 创建的 Map 每隔 spec 会随机检查 deleteScale 的数据，如果检查到的项目过期
// 他们会被删除,如果不希望主动检查，请使用 NewWithLazyDelete
//
// deleteScale 应该是大于 0 小于 1 的小数，否则的话将使用默认值 0.5
func NewBaseAgingMap(spec time.Duration, deleteScale float64) *AgingMap {
	if deleteScale > 1 || deleteScale <= 0 {
		deleteScale = 0.5
	}
	am := &AgingMap{
		_map:        Map{},
		task:        cron.New(),
		deleteScale: deleteScale,
	}
	sec := fmt.Sprintf("*/%d * * * * ?", int(spec.Seconds()))
	_ = am.task.AddFunc(sec, am.deleteExpiredItems)
	am.task.Start()
	return am
}

// NewWithLazyDelete 用来创建一个惰性删除的时效 Map.
// 惰性删除是指只有在进行 Load 操作时 againMap 才会去判断某一项有没有过期，
// 如果过期了，它会被删除，所以如果某一项一直没有被读取，那他将永远不会被删除。
// 与之对应的是使用 NewAgingMap 创建 Map, 这类 Map 会有定时任务定时清理已经过期的项。
func NewWithLazyDelete() *AgingMap {
	return &AgingMap{
		_map: Map{},
	}
}

// Store 用于向 Map 中存入一条数据，如果 key 已经存在，旧值将被新值覆盖。
// age 用于指定该键值对的生存时长。
func (am *AgingMap) Store(key, v interface{}, age time.Duration) {
	am._map.Store(key, agingValue{
		v:          v,
		age:        age,
		createTime: time.Now().UnixNano(),
	})
}

// Load 类似于 LoadWithDeadline，但他舍弃了 deadline
func (am *AgingMap) Load(key interface{}) (val interface{}, ok bool) {
	val, _, ok = am.LoadWithDeadline(key)
	return
}

// LoadWithDeadline 根据 key 返回三个值：val 表示 key 对应的 value,
// deadline 代表该键值对剩余的生存时间，单位秒，ok 表示 key 是否存在
// 如果 key 不存在，将会返回 nil, 0, false
func (am *AgingMap) LoadWithDeadline(key interface{}) (val interface{}, deadline float64, ok bool) {
	now := time.Now().UnixNano()
	v, o := am._map.Load(key)
	if !o {
		return nil, 0, false
	}
	av := v.(agingValue)
	if now-av.createTime > av.age.Nanoseconds() {
		am.Delete(key)
		return nil, 0, false
	}
	d := time.Duration(av.age.Nanoseconds() - (now - av.createTime))
	sec := d / time.Second
	nsec := d % time.Second
	return av.v, float64(sec) + float64(nsec)/1e9, true
}

// Delete 用于删除 key 对应的键值对，不管他有没有过期。
func (am *AgingMap) Delete(key interface{}) {
	am._map.Delete(key)
}

// Range 用来遍历 Map 中的键值对，遍历到的 k, v 将被赋值给 f 的两个参数
// f 返回 false 时，遍历会结束，使用方法如下：
//  // am := NewAgingMap()
//  // // ...
//  // am.Range(func(k, v interface{}) bool {
//  //     fmt.Println(k, v)
//  //     return true
//  // }
func (am *AgingMap) Range(f func(k, v interface{}) bool) {
	am._map.Range(func(key, value interface{}) bool {
		val := value.(agingValue)
		if time.Now().UnixNano()-val.createTime > val.age.Nanoseconds() {
			am.Delete(key)
			return true
		}
		return f(key, value.(agingValue).v)
	})
}
