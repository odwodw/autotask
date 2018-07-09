package main

import (
	"fmt"
	"sync/atomic"
	"time"
	"github.com/odwodw/autotask"
	"reflect"
)

var sum int32 = 0
var tt *gotimer.Timer
var tmmap map[string] *(gotimer.Task)

func nowtime(sum int) {
	NOW := time.Now()
	fmt.Println(reflect.TypeOf(NOW), NOW.Format("2006-01-02 15:04:05."),NOW.Nanosecond()/1e3,sum)
}

func now() {
	atomic.AddInt32(&sum, 1)
	nowtime(int(sum))
	v := atomic.LoadInt32(&sum)
	
	node, ok := tmmap["auto_report_time"]
	if !ok {
			return
	}
	switch v {
		case 5:
			fmt.Println("Timer Pause...")
			node.Pause()
		case 8:
			fmt.Println("Timer Stop...")
			node.Stop()
			
	}
}

func sleepExec(secs time.Duration, f func(args interface{}), args interface{}) {
	time.Sleep(time.Second*secs)
	fmt.Println(f)
	f(args)
}

func main() {
	tt := gotimer.NewScheduler()
	fmt.Println(tt)
	
	tmmap = make(map[string] *gotimer.Task)
	atn := gotimer.NewTask("auto_report_time").Every(1).Do(now).Once().WorkTime("08:00 - 18:00").At("2018-07-09 18:44:40")
	//atn := gotimer.NewTask("auto_report_time")
	fmt.Println(atn)
	tmmap["auto_report_time"] = atn
	/*
	go func() {
		time.Sleep(time.Second*10)
		fmt.Println("Timer Resume...")
		atn.Resume()
	}()		

	go func() {
		time.Sleep(time.Second*17)
		fmt.Println("NewAtTimer Start...")
		atn.AddToTimer(tt)
	}()		
	go func() {
		time.Sleep(time.Second*20)
		fmt.Println("All Stop...")
		tt.Stop()
	}()
	*/
	
	nowtime(0)
	
	//time.Sleep(time.Second*3)
	atn.AddToTimer(tt)
	tt.Start()
	for {	
		time.Sleep(time.Hour*3)
	}
}
