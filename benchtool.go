package benchtool

import (
	"fmt"
	"sync/atomic"
	"time"
)

// BenchTask 压力测试任务
type BenchTask interface {
	InitData(num int) ([]interface{}, error)
	Exec(data interface{}) error
}

// BenchTooler 压力测试工具
type BenchTooler interface {
	Start(connectnum, requestnum int, task BenchTask) error
}

var defaultBenchTool BenchTooler

func init() {
	defaultBenchTool = CreateBenchTool()
}

// Start 开始压测任务
func Start(connectnum, requestnum int, task BenchTask) error {
	return defaultBenchTool.Start(connectnum, requestnum, task)
}

// CreateBenchTool 创建压力测试工具
func CreateBenchTool() BenchTooler {
	return &benchTool{}
}

type benchTool struct {
	task       BenchTask
	connectnum int
	requestnum int
	taskchan   chan interface{}
	successNum int64
	errorNum   int64

	startTime time.Time
}

func (b *benchTool) Start(connectnum, requestnum int, task BenchTask) error {
	b.task = task
	b.connectnum = connectnum
	b.requestnum = requestnum
	b.taskchan = make(chan interface{}, 1024)
	b.runInitExec()
	err := b.runInitData()
	if err != nil {
		return err
	}
	b.runShowStaus()
	t := time.Since(b.startTime)
	fmt.Printf("[%s]本次压测共用时:%v 平均每秒次数:%f\n", time.Now().Format("15:04:05.000"), t, float64(requestnum)/float64(t/time.Second))
	return nil
}

func (b *benchTool) runShowStaus() {
	b.showStatus(int64(b.requestnum))
}
func (b *benchTool) runInitExec() {
	connectnum := b.connectnum
	for i := 0; i < connectnum; i++ {
		go func(i int) {
			for data := range b.taskchan {

				err := b.task.Exec(data)
				if err != nil {
					b.addErr(err)
				} else {
					b.addSuccess()
				}
			}
		}(i)
	}
}
func (b *benchTool) runInitData() error {
	data, err := b.task.InitData(b.requestnum)
	if err != nil {
		return err
	}
	b.startTime = time.Now()
	for _, d := range data {
		b.taskchan <- d
	}
	return nil
}
func (b *benchTool) showStatus(sum int64) {
	for sum > b.successNum+b.errorNum {
		time.Sleep(time.Second)
		fmt.Printf("\r[%s]成功提交:\033[1;32m%d\033[0m  出错:\033[1;31m%d\033[0m", time.Now().Format("15:04:05.000"), b.successNum, b.errorNum)
	}
	fmt.Println()
}
func (b *benchTool) addErr(msg error) {
	atomic.AddInt64(&b.errorNum, 1)
	fmt.Printf("[%s]出错了:%v\n", time.Now().Format("15:04:05.000"), msg)
}
func (b *benchTool) addSuccess() {
	atomic.AddInt64(&b.successNum, 1)
}
