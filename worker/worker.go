package worker

// 任务
type Job interface {
    Do()
}

// 工作通道
type JobChannel chan Job

// 工作者
type Worker struct {
    workerPool chan JobChannel // 工作池，存放工作通道
    jobChannel JobChannel      // 工作通道，存放任务队列
    quit       chan bool       // 退出信号
}

// 新建一个工作者
func NewWorker(workerPool chan JobChannel) *Worker {
    return &Worker{
        workerPool: workerPool,
        jobChannel: make(JobChannel),
        quit:       make(chan bool),
    }
}

// 启动工作通道
func (w *Worker) Start() {
    go func() {
        for {
            // 注册空闲的工作通道到工作池
            w.workerPool <- w.jobChannel
            select {
            // 执行任务
            case job := <-w.jobChannel:
                if job != nil {
                    job.Do()
                }
            // 停止工作通道
            case <-w.quit:
                return
            }
        }
    }()
}

// 停止工作通道
func (w *Worker) Stop() {
    w.quit <- true
}
