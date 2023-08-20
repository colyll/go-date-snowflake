# go-date-snowflake
Add date string before at Snowflake ID's string

```
/**
* 由于雪花算法加上日期有26位长，所以修改缩小到22位，容量至9999-12-31，104.8W/ms。
* |-------------------------------雪花算法(64bits)----------------------------------------|
* |---补位(1bit)--|---当前时间戳毫秒(41bits)----|--机器ID(10bits)--|--序号(12bits)--|
*
*          |---------------------------------修改后(48bits)--------------------------------|
*  日期 + |--补位(1bit)--|--每日第几毫秒(27bits)--|--机器ID(9bits)--|-序号(11bits)--|
*
* e.g 2023081838643360073728
  */
```

## 使用方法
配置文件`config.yaml`, 置于运行目录下, 或者自行修改读取的位置。

使用redis来缓存自增序号, 相比使用sync包方案, 
虽增加了消耗, 但也减少了重复执行同任务且未修改配置带来的风险。

可以设置数据中心区段，6bit的情况下，最后生成字符串长24位

```go
package main

import (
	"fmt"
	"github.com/colyll/go-date-snowflake"
	"github.com/redis/go-redis/v9"
	"time"
)

func main() {
	t1 := time.Now()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	s := snowflake.New(client)

	for i := 0; i < 100000; i++ {
		_, err := s.Id()
		if err != nil {
			fmt.Println("error!")
		}
	}
	t2 := time.Since(t1)

	fmt.Println(t2.Microseconds())
}
```
