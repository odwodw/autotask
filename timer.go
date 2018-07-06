package gotimer

import (
	"container/list"
	"fmt"
	"sync"
	"time"
	"reflect"
)

//referer https://github.com/cloudwu/skynet/blob/master/skynet-src/skynet_timer.c

const (
	TIME_NEAR_SHIFT  = 8
	TIME_NEAR        = 1 << TIME_NEAR_SHIFT
	TIME_LEVEL_SHIFT = 6
	TIME_LEVEL       = 1 << TIME_LEVEL_SHIFT
	TIME_NEAR_MASK   = TIME_NEAR - 1
	TIME_LEVEL_MASK  = TIME_LEVEL - 1
	TIMER_RESUME = '0'
	TIMER_PAUSE = '1'
	TIMER_STOP = '2'
)

type Timer struct {
	near [TIME_NEAR]*list.List
	t    [4][TIME_LEVEL]*list.List
	sync.Mutex
	time uint32
	Tick time.Duration
	quit chan struct{}
}

type CheckFunc func(*Timer) bool

type TimerNode struct {
	name string
	sync.Mutex
	starttime time.Time //开始执行时间
	lasttime time.Time  //最后一次执行时间
	times uint32	    //已执行次数
	operflag byte       //操作标志TIMER_RESUME:恢复;TIMER_PAUSE:暂停;TIMER_STOP:停止;
	unit time.Duration  //执行间隔时间单位
	interval uint32     //执行间隔
	UserCheck CheckFunc//用户定义检查函数,供二次开发
	expire uint32
	taskf interface{}       //任务函数
	fParams []interface{} //任务函数f的参数列表
}

func (n *TimerNode)defaultUserCheck(t *Timer) bool {
	return true
}

func defaultaskf() {
	fmt.Println("task func noset!!!")
}

/*
func (n *TimerNode) String() string {
	return fmt.Sprintf("TimerNode:name:%s,interval:%d,times:%d", n.name, n.interval, n.times)
}*/

/*taskfunc run*/
func (n *TimerNode) Run() {
	fn := reflect.ValueOf(n.taskf)

	params := make([]reflect.Value, len(n.fParams))
	for key, value := range n.fParams {
		params[key] = reflect.ValueOf(value)
	}

	fn.Call(params)
}

/*set taskfunc*/
func (n *TimerNode) Do(taskf interface{}, params ...interface{}) *TimerNode {
	n.taskf = taskf
	n.fParams = params
	return n
}

//设置间隔
func (n *TimerNode) Every(interval uint32) *TimerNode {
	n.interval = interval
	return n
}

//stop
func (n *TimerNode) Stop() {
	n.Lock()
	defer n.Unlock()
	n.operflag = TIMER_STOP
}
//pause
func (n *TimerNode) Pause() {
	n.Lock()
	defer n.Unlock()
	n.operflag = TIMER_PAUSE
}
//resume
func (n *TimerNode) Resume() {
	n.Lock()
	defer n.Unlock()
	n.operflag = TIMER_RESUME
}

//addtoTimer
func (n *TimerNode) AddToTimer(t *Timer) {
	t.NewTimerWithNode(n)
}

//readdtoTimer
func (n *TimerNode) ReAddToTimer(t *Timer) {
	t.Lock()
	defer t.Unlock()
	n.expire = n.interval + t.time
	t.AddTimerNode(n, false)
}

//newnode
func NewTimerNode(name string) *TimerNode {
	n := new(TimerNode)
	n.name = name
	n.UserCheck = n.defaultUserCheck
	n.taskf = defaultaskf
	return n
}

//newtimer
func (t *Timer) NewNode(name string, interval uint32, tf interface{}, tfparams ...interface{}) *TimerNode {
	n := NewTimerNode(name).Every(interval).Do(tf, tfparams)
	n.starttime = time.Now()
	n.lasttime = n.starttime
	n.operflag = TIMER_RESUME
	t.Lock()
	defer t.Unlock()
	n.expire = n.interval + t.time
	t.AddTimerNode(n,false)
	return n
}
//newtimerwithnode
func (t *Timer) NewTimerWithNode(n *TimerNode) {
	n.starttime = time.Now()
	n.lasttime = n.starttime
	n.operflag = TIMER_RESUME
	t.Lock()
	defer t.Unlock()
	n.expire = n.interval + t.time
	t.AddTimerNode(n,false)
}

func NewTimer(d time.Duration) *Timer {
	t := new(Timer)
	t.time = 0
	t.Tick = d
	t.quit = make(chan struct{})

	var i, j int
	for i = 0; i < TIME_NEAR; i++ {
		t.near[i] = list.New()
	}

	for i = 0; i < 4; i++ {
		for j = 0; j < TIME_LEVEL; j++ {
			t.t[i][j] = list.New()
		}
	}

	return t
}

func (t *Timer) AddTimerNode(n *TimerNode, first bool) {
	if first {
		t.near[t.time&TIME_NEAR_MASK].PushBack(n)
		return
	}
	expire := n.expire
	current := t.time
	if (expire | TIME_NEAR_MASK) == (current | TIME_NEAR_MASK) {
		t.near[expire&TIME_NEAR_MASK].PushBack(n)
	} else {
		var i uint32
		var mask uint32 = TIME_NEAR << TIME_LEVEL_SHIFT
		for i = 0; i < 3; i++ {
			if (expire | (mask - 1)) == (current | (mask - 1)) {
				break
			}
			mask <<= TIME_LEVEL_SHIFT
		}
		t.t[i][(expire>>(TIME_NEAR_SHIFT+i*TIME_LEVEL_SHIFT))&TIME_LEVEL_MASK].PushBack(n)
	}

}

func (t *Timer) String() string {
	return fmt.Sprintf("Timer:time:%d, tick:%s", t.time, t.Tick)
}

//检查定时器节点执行条件
func (t *Timer) checkNode(n *TimerNode) bool {
	n.Lock()
	defer n.Unlock()
	nowtime :=  time.Now()
	fmt.Println("sub:", nowtime.Sub(n.lasttime))
	n.lasttime = nowtime
	if n.operflag == TIMER_STOP {
		return false
	}
	
	if !n.UserCheck(t) {
		return false
	}
	
	if n.operflag == TIMER_RESUME {
		n.times++
		return true
	}
	return false
}

func (t *Timer) dispatchList(front *list.Element) {
	for e := front; e != nil; e = e.Next() {
		node := e.Value.(*TimerNode)
		if t.checkNode(node) {
			go node.Run()
		}
	}
}

func (t *Timer) moveList(level, idx int) {
	vec := t.t[level][idx]
	front := vec.Front()
	vec.Init()
	for e := front; e != nil; e = e.Next() {
		node := e.Value.(*TimerNode)
		t.AddTimerNode(node,false)
	}
}

func (t *Timer) shift() {
	t.Lock()
	defer t.Unlock()
	var mask uint32 = TIME_NEAR
	t.time++
	ct := t.time
	if ct == 0 {
		t.moveList(3, 0)
	} else {
		time_shift := ct >> TIME_NEAR_SHIFT
		var i int = 0
		for (ct & (mask - 1)) == 0 {
			idx := int(time_shift & TIME_LEVEL_MASK)
			if idx != 0 {
				t.moveList(i, idx)
				break
			}
			mask <<= TIME_LEVEL_SHIFT
			time_shift >>= TIME_LEVEL_SHIFT
			i++
		}
	}
}

func (t *Timer) execute() {
	t.Lock()
	idx := t.time & TIME_NEAR_MASK
	vec := t.near[idx]
	if vec.Len() > 0 {
		front := vec.Front()
		vec.Init()
		t.Unlock()
		// dispatch_list don't need lock
		t.dispatchList(front)
		return
	}
	t.Unlock()
}

func (t *Timer) update() {
	// try to dispatch timeout 0 (rare condition)
	t.execute()

	// shift time first, and then dispatch timer message
	t.shift()

	t.execute()

}

func (t *Timer) Start() {
	go func() {
		tick := time.NewTicker(t.Tick)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				t.update()
			case <-t.quit:
				return
			}
		}
	}()
}

func (t *Timer) Stop() {
	close(t.quit)
}
