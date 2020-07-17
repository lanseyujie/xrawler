package worker

// 任务调度器
type Dispatcher struct {
    workerPool chan JobChannel // 工作池，存放工作通道
    jobQueue   JobChannel      // 任务队列
    maxWorkers int             // 最多工作者
    workers    []*Worker       // 保存工作者
    quit       chan bool       // 退出信号
}

// 新建一个任务调度器
func NewDispatcher(maxWorkers, maxJobs int) *Dispatcher {
    return &Dispatcher{
        workerPool: make(chan JobChannel, maxWorkers),
        jobQueue:   make(JobChannel, maxJobs),
        maxWorkers: maxWorkers,
        workers:    make([]*Worker, 0, maxWorkers),
        quit:       make(chan bool),
    }
}

// 注册工作通道到工作池并调度任务
func (d *Dispatcher) Start() {
    // 创建工作者
    for i := 0; i < d.maxWorkers; i++ {
        worker := NewWorker(d.workerPool)
        d.workers = append(d.workers, worker)
        // 注册空闲工作通道并等待任务执行
        worker.Start()
    }

    // 分发调度新任务到工作通道
    go func() {
        for {
            select {
            // 分发新任务
            case job := <-d.jobQueue:
                go func() {
                    // 取出一个空闲的工作通道
                    jobChannel := <-d.workerPool
                    // 将任务加入该工作通道
                    jobChannel <- job
                }()
            // 停止分发任务
            case <-d.quit:
                return
            }
        }
    }()
}

// 将任务加入队列
func (d *Dispatcher) Put(job Job) {
    d.jobQueue <- job
}

// 停止分发任务
func (d *Dispatcher) Stop() {
    d.quit <- true
}

// 停止执行任务
func (d *Dispatcher) StopWorker() {
    d.quit <- true
    for _, worker := range d.workers {
        go worker.Stop()
    }
}
