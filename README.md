# go-executors
go 轻量线程池

## usage:

```
go get -u github.com/longmon/go-executors
```

```go
//全局初始化线程池，应仅在启动时执行一次，多次执行无效
//初始化100个线程，执行过程动态调整线程，最多1000个，最少100
executors.InitExecutorWithCapacity(100, 1000)

//添加异步执行任务， 返回任务对象和错误，
//支持三种类型的闭包 `func()`, `func() error`,`func()(interface{}, error)`
job, err := executors.Run(func(){
    //task code here
})

//job有两个公开的字段，job.Err和job.Result 分别对应闭包的返回类型
//至于是不是每次都有值，那就要看你传入的闭包类型了

//仅在线程池关闭后添加任务会返回错误
if err != nil {
    log.Fatalln(err)
}

//同步等待任务执行结束，返回的错误信息是任务执行过程中产生的panic
//如果不关注返回或不必等待执行结果可不处理返回
err = job.Wait(func(){
  //run after task done
})

//优雅地关闭线程池，一般情况下不必关闭吧
executors.Shutdown()


```
