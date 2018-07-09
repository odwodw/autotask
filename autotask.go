package gotimer

import (
	"fmt"
	"time"
	"strings"
	"strconv"
)

type Task struct {
	once bool    		//不显式设置,once默认是循环执行(false);
	worktime WorkTime	//每日工作时间区间
	attime time.Time	//计划执行时间,最低单位至秒
	beginday time.Time	//工作起始日
	endday time.Time	//工作终止日
	TimerNode
}

func NewTask(name string) *Task {
	t := new(Task)
	t.name = name	
	t.UserCheck = t.myCheckFunc	
	t.taskf = defaultaskf
	return t
}

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

//检查当前时间是否在工作时间范围
func (wt *WorkTime)Check(currtime time.Time) bool {
	nhour := currtime.Hour()
	nminute := currtime.Minute()
	if nhour < wt.beginh || nhour > wt.endh {
		return false
	}
	if (nhour == wt.beginh && nminute < wt.beginm) || (nhour == wt.endh && nminute > wt.endm) {
		return false
	}
	return true
}

/*根据Task中新增的属性判断是否执行任务*/
func (t *Task)myCheckFunc(tt *Timer) bool{

	//是循环任务就重新添加定时器;
	if !t.once {
		t.ReAddToTimer(tt)
	}
	
	//是否已过计划执行时间
	if t.lasttime.Before(t.attime) {
		if t.once {/*计划时间点任务一次性的也要重新添加*/
			t.ReAddToTimer(tt)
		}
		fmt.Println("attime Check triggered(false)")
		return false
	}
	
	//是否在工作时间区间
	if !t.worktime.Check(t.lasttime) {
		fmt.Println("worktime Check triggered(false)")
		return false
	}
	
	return true
}

func (t *Task)Once() *Task {
	t.once = true
	return t
}

func (t *Task)Cycle() *Task {
	t.once = false
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
/*set attime
 *ats-计划时间字符串, 格式:YYYY-MM-DD HH:mm:ss
 */
func (t *Task) At(ats string) *Task {
	tokens := strings.FieldsFunc(ats, 
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
	fmt.Println(ats,"len", len(tokens))
	if len(tokens) == 6 {
		year,_ := strconv.Atoi(tokens[0])
		month,_ := strconv.Atoi(tokens[1])
		day,_ := strconv.Atoi(tokens[2])
		hour,_ := strconv.Atoi(tokens[3])
		min,_ := strconv.Atoi(tokens[4])
		sec,_ := strconv.Atoi(tokens[5])
		t.attime = time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
		return t
	}
	
	return t
}

