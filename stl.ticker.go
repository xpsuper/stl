package stl

import "time"
import "sync"

/*** 使用示例 ****

func testFn(args interface{}) {
	Println(args)
}

func main() {

	task := NewTicker(100, true) //100毫秒扫描一次任务，任务加入时执行一次
	defer task.Stop()

	ch2 := make(chan bool, 0)
	ch4 := make(chan bool, 0)
	id1 := task.AddFunc(test, 3, "任务111111")                // 每间隔3秒执行一次 test 函数
	id2 := task.AddChannel(ch2, 5)                           // 每间隔5秒向ch2回写 bool true
	id3 := task.AddTimedFunc(test, "16:18:55", "任务333333")  // 每天 16:18:55 执行一次 test，如果在建立该任务时当前时间已经超过该指定时间，则任务推迟至明天同一时间执行
	id4 := task.AddTimedChannel(ch4, "16:18:55")             // 每天 16:18:55 向ch4回写 bool true，如果在建立该任务时当前时间已经超过该指定时间，则任务推迟至明天同一时间执行

	task.Cancel(id1)                                         // 取消任务 id1

	Println(id1, id2, id3, id4)
	for {
		select {
		case <- ch2:
			Println("任务222222")
		case <- ch4:
			Println("任务444444")
		}
	}
}
 */

type TickerTasks struct {
	exec bool // 是否在打开任务时执行一次
	exit bool
	ChanClosed chan bool
	sync.Mutex
	maxTaskId int
	scanInterval int // 每间隔 scanInterval 毫秒扫一次任务
	taskList []*TickerTask
}

type TickerTask struct {
	id 		      int
	intervalTime  int64 // 任务间隔时间
	lastHnTime    int64 // 上一次执行时间
	taskType      uint8 // 1-普通周期性任务 2-定时周期性任务
	timerTime     time.Time // 针对定时周期性任务时使用
	callbackChan  chan <- bool
	taskParams    interface{}
	handleFunc    func(arg interface{})
}

func NewTicker(scanInterval int, execOnStart bool) *TickerTasks {
	if scanInterval < 50 {
		scanInterval = 50
	}
	ticker := &TickerTasks {
		scanInterval: scanInterval,
		exec: execOnStart,
		ChanClosed: make(chan bool, 1),
	}
	go ticker.listenTickerTasks()
	return ticker
}

func (ticker *TickerTasks) AddChannel(ch chan <- bool, interval int64) int {
	ticker.Lock()
	defer ticker.Unlock()

	task := &TickerTask {
		id: ticker.maxTaskId,
		taskType: 1, // 普通周期性任务，如：每隔 10 秒运行一次的任务
		intervalTime: interval,
		callbackChan: ch,
	}
	if !ticker.exec {
		task.lastHnTime = time.Now().Unix()
	}
	ticker.maxTaskId++
	ticker.taskList = append(ticker.taskList, task)
	return ticker.maxTaskId
}

func (ticker *TickerTasks) AddFunc(fn func (arg interface{}), interval int64, params interface{}) int {
	ticker.Lock()
	defer ticker.Unlock()

	ticker.maxTaskId++
	task := &TickerTask {
		id: ticker.maxTaskId,
		taskType: 1, // 1-指定为 ticker 周期性任务, 如：每隔 10 秒运行一次的任务
		intervalTime: interval,
		handleFunc: fn,
		taskParams: params,
	}
	if !ticker.exec {
		task.lastHnTime = time.Now().Unix()
	}
	ticker.taskList = append(ticker.taskList, task)
	return ticker.maxTaskId
}

func (ticker *TickerTasks) AddTimedFunc(k func (arg interface{}), ts string, params interface{}) int {
	ticker.Lock()
	defer ticker.Unlock()

	nt := time.Now().UTC()
	sr := stringToTime(nt.Format("2006-01-02") + " " + ts)
	if !sr.After(nt) {
		sr = sr.AddDate(0, 0, 1) // 推迟一天执行
	}
	ticker.maxTaskId++

	ticker.taskList = append(ticker.taskList, &TickerTask {
		id: ticker.maxTaskId,
		taskType: 2, // 2-指定为 timer 定时周期性任务, 如：每天5:00:00运行一次的任务
		timerTime: sr,
		handleFunc: k,
		taskParams: params,
	})
	return ticker.maxTaskId
}

func (ticker *TickerTasks) AddTimedChannel(ch chan <- bool, ts string) int {
	ticker.Lock()
	defer ticker.Unlock()

	nt := time.Now().UTC()
	sr := stringToTime(nt.Format("2006-01-02") + " " + ts)
	if !sr.After(nt) {
		sr = sr.AddDate(0, 0, 1) // 推迟一天执行
	}
	ticker.maxTaskId++
	ticker.taskList = append(ticker.taskList, &TickerTask {
		id: ticker.maxTaskId,
		taskType: 2, // 2-指定为 timer 定时周期性任务, 如：每天5:00:00运行一次的任务
		timerTime: sr,
		callbackChan: ch,
	})
	return ticker.maxTaskId
}

// 传入任务 id 以取消该任务
func (ticker *TickerTasks) Cancel(taskId int) {
	ticker.Lock()
	defer ticker.Unlock()
	for i, task := range ticker.taskList {
		if task.id == taskId {
			ticker.taskList = append(ticker.taskList[:i], ticker.taskList[i+1:]...)
		}
	}
}

func (ticker *TickerTasks) Stop() {
	ticker.exit = true
}

func (ticker *TickerTasks) listenTickerTasks() {
	for {
		if ticker.exit {
			ticker.ChanClosed <- true
			println("exit for running task...")
			return
		}
		// 周期性任务
		now := time.Now().UTC()
		for _, task := range ticker.taskList {
			handle := false
			if task.taskType == 1 && now.Unix() - task.lastHnTime + 1 > task.intervalTime {
				handle = true
				task.lastHnTime = now.Unix()
			} else if task.taskType == 2 && now.After(task.timerTime) {
				handle = true
				task.timerTime = task.timerTime.AddDate(0, 0, 1)
			}
			if handle {
				if task.handleFunc != nil {
					go task.handleFunc(task.taskParams)
				} else {
					go func(t *TickerTask) {
						select {
						case t.callbackChan <- true:
						case <- time.After(time.Second):
							println("time out...")
						}
					}(task)
				}
			}
		}
		time.Sleep(time.Millisecond * time.Duration(ticker.scanInterval))
	}
}

func stringToTime(dateTime string) time.Time {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, dateTime, loc)
	return theTime
}