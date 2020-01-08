

## 介绍

dispatcher 是 Go 语言实现的作业调度工具包。它提供了一种简单、人性化的方式去调度 Go 函数，包括延迟和周期性两种调度方式。

灵感来源于 Linux [cron](https://opensource.com/article/17/11/how-use-cron-linux) 和 Python [schedule](https://github.com/dbader/schedule)。

## 功能

- 延迟执行，精确到一秒钟
- 周期性执行，精确到一秒钟，类似 cron 的风格，但是更加的灵活
- 取消 job
- 失败重试（暂时重试一次）

## 安装

```go
go get github.com/xpsuper/stl/dispatcher
```

## 例子

job 函数

```Go
func task1(name string, age int) {
	fmt.Printf("run task1, with arguments: %s, %d\n", name, age)
}

func task2() {
	fmt.Println("run task2, without arguments")
}
```

### 延迟调度

延迟调度支持四种模式：按秒、分、小时、天。

作为特例，任务将通过 `s.Delay().Do(task)` 立即执行。

```Go
package main

import (
    "fmt"

    "github.com/xpsuper/stl/dispatcher"
)

func main() {
	s, err := dispatcher.NewDispatcher(1000)
	if err != nil {
		panic(err) // just example
	}

	// delay with 1 second, job function with arguments
	s.Delay().Second(1).Do(task1, "prprprus", 23)

	// delay with 1 minute, job function without arguments
	s.Delay().Minute(1).Do(task2)

	// delay with 1 hour
	s.Delay().Hour(1).Do(task2)

	// special: execute immediately
	s.Delay().Do(task2)

	// cancel job
	jobID := s.Delay().Day(1).Do(task2)
	err = s.CancelJob(jobID)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("cancel delay job success")
	}
}
```

### 周期性调度

类似 cron 的风格，同样会包括秒、分、小时、天、星期、月，但是它们之间的顺序和数量不需要固定成一个死格式。你可以按照你的个人喜好去进行排列组合。例如，`Second(3).Minute(35).Day(6)` 和 `Minute(35).Day(6).Second(3)` 的效果是一样的。不需要再去记格式了！🎉👏

但是为了可读性，推荐按照从小到大（或者从大到小）的顺序使用。

注意：`Day()` 和 `Weekday()` 避免同时出现，除非你清楚知道这天是星期几。

作为特例，任务将通过 `s.Every().Do(task)` 每秒被执行一次。

```Go
package main

import (
    "fmt"

    "github.com/xpsuper/stl/dispatcher"
)

func main() {
	s, err := dispatcher.NewDispatcher(1000)
	if err != nil {
		panic(err)
	}

	// Specifies time to execute periodically
	s.Every().Second(45).Minute(20).Hour(13).Day(23).Weekday(3).Month(6).Do(task1, "prprprus", 23)
	s.Every().Second(15).Minute(40).Hour(16).Weekday(4).Do(task2)
	s.Every().Second(1).Do(task1, "prprprus", 23)

	// special: executed once per second
	s.Every().Do(task2)

	// cancel job
	jobID := s.Every().Second(1).Minute(1).Hour(1).Do(task2)
	err = s.CancelJob(jobID)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("cancel periodically job success")
	}
}
```
