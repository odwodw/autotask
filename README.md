# autotask
golang实现的自动任务库

基于skynet的时间轮timer的golang实现

# 主要功能

* 可添加两种精度调度器(按秒、按天)(时间轮精度)

* 任务的添加、暂停、恢复、清除(停止)

* 任务可设置一次性、循环执行

* 任务可设置间隔执行、计划时间点(YYYY-MM-DD HH:mm:ss)执行

* 任务可设置工作时间区间(HH:mm - HH:mm)


# 安装方法

```bash
	go get github.com/odwodw/autotask
```

# 使用方法

* 基本使用

```bash
tt := gotimer.NewScheduler()
atn := gotimer.NewTask("auto_report_time").Every(1).Do(funcname).Once().WorkTime("08:00 - 18:00").At("2018-07-09 18:44:40")
atn.AddToTimer(tt)
tt.Start()
```

# 备注

