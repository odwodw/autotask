package gotimer

import (
	"time"
	"strings"
	"strconv"
)

//以秒为单位的调度器
func NewScheduler() *Timer {
	t := NewTimer(time.Second)
	return t
}
//以天为单位的调度器
func NewSchedulerWithDay() *Timer {
	t := NewTimer(24*time.Hour)
	return t
}

//每日工作时间区间定义
type WorkTime struct {
	beginh, beginm, endh, endm int
}

//分隔形如"00:00-23:59"的字符串
func (wt *WorkTime)SetString(wts string) *WorkTime{
	tokens := strings.FieldsFunc(wts, 
		func(s rune) bool {
			switch s {
				case ' ':
					return true
				case ':':
					return true
				case '-':
					return true
			}
			return false
		})
	if len(tokens) < 4 {
		wt.beginh = 0
		wt.beginm = 0
		wt.endh = 23
		wt.endm = 59
	} else {
		wt.beginh,_ = strconv.Atoi(tokens[0])
		wt.beginm,_ = strconv.Atoi(tokens[1])
		wt.endh,_ = strconv.Atoi(tokens[2])
		wt.endm,_ = strconv.Atoi(tokens[3])
	}
	return wt
}

type Task struct {
	once bool
	worktime WorkTime	//每日工作时间区间
	attime time.Time    //计划执行时间,最低单位至秒
	beginday time.Time  //工作起始日
	endday time.Time    //工作终止日
	TimerNode
}

func NewTask(name string) *Task {
	t := new(Task)
	t.name = name	
	t.UserCheck = t.myCheckFunc	
	t.taskf = defaultaskf
	return t
}

/*根据Task中新增的属性判断是否执行任务*/
func (t *Task)myCheckFunc(tt *Timer) bool{
	//是循环任务就重新添加定时器
	if !t.once {
		t.ReAddToTimer(tt)
	}
	
	return true
}

func (t *Task)Once() *Task {
	t.once = true
	return t
}

/*reload func*/
func (t *Task) Do(taskf interface{}, params ...interface{}) *Task {
	t.taskf = taskf
	t.fParams = params
	return t
}

/*reload func*/
func (t *Task) Every(interval uint32) *Task {
	t.interval = interval
	return t
}

/*set worktime*/
func (t *Task) WorkTime(wts string) *Task {
	wt := new(WorkTime)
	wt.SetString(wts)
	t.worktime = *wt
	return t
}
/*set attime*/
func (t *Task) At(wts string) *Task {
	wt := new(WorkTime)
	wt.SetString(wts)
	t.worktime = *wt
	return t
}

