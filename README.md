# go-executors
go 轻量线程池

## usage:

```go
//全局初始化线程池，应仅在启动时执行一次，多次执行无效
//初始化100个线程，执行过程动态调整线程，最多1000个，最少100
executors.InitExecutorWithCapacity(100, 1000)

//添加异步执行任务， 返回任务对象和错误，
job, err := executors.Run(func(){
    //task code here
})

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
