package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"sync"
	"time"
)

// Worker 工作池中的工作线程
// 负责处理消息请求
type Worker struct {
	WorkerID     uint32
	JobQueue     chan zinterface.IRequest
	WorkerPool   chan chan zinterface.IRequest
	isStopped    bool
	stopChan     chan bool
	lastActivity time.Time
	isCore       bool
	wg           sync.WaitGroup
}

// NewWorker 创建新的工作线程
func NewWorker(workerID uint32, pool chan chan zinterface.IRequest, isCore bool) *Worker {
	return &Worker{
		WorkerID:     workerID,
		JobQueue:     make(chan zinterface.IRequest),
		WorkerPool:   pool,
		isStopped:    false,
		stopChan:     make(chan bool),
		lastActivity: time.Now(),
		isCore:       isCore,
	}
}

// Start 启动工作线程
func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			// 将工作线程的JobQueue注册到WorkerPool中
			w.WorkerPool <- w.JobQueue

			select {
			// 接收任务
			case request := <-w.JobQueue:
				// 更新最后活动时间
				w.lastActivity = time.Now()
				// 处理消息请求
				request.GetConnection().GetRouter().DoMsgHandler(request)
			// 接收停止信号
			case <-w.stopChan:
				w.isStopped = true
				return
			}
		}
	}()
}

// Stop 停止工作线程
func (w *Worker) Stop() {
	if w.isStopped {
		return
	}
	w.isStopped = true
	close(w.stopChan)
	w.wg.Wait()
}

// WorkerPool 工作池
// 管理一组工作线程，用于处理消息请求
type WorkerPool struct {
	// 核心工作线程数
	coreWorkers uint32
	// 最大工作线程数
	maxWorkers uint32
	// 请求队列大小
	queueSize uint32
	// 非核心工作线程空闲超时时间（秒）
	idleTimeout time.Duration
	// 当前工作线程数
	currentWorkers uint32
	// 工作线程池
	WorkerPool chan chan zinterface.IRequest
	// 请求队列
	JobQueue chan zinterface.IRequest
	// 工作线程集合
	workers map[uint32]*Worker
	// 互斥锁，保护工作池的并发访问
	mutex sync.RWMutex
	// 停止通道
	stopChan chan bool
	// 是否已停止
	isStopped bool
	// 等待组
	wg sync.WaitGroup
	// 下一个工作线程ID
	nextWorkerID uint32
	// 任务调度器是否已启动
	dispatcherStarted bool
	// 定时器，用于检查空闲工作线程
	timer *time.Timer
	// 定时器是否已启动
	timerStarted bool
}

// NewWorkerPool 创建新的工作池
// 使用默认配置
func NewWorkerPool() *WorkerPool {
	return NewWorkerPoolWithConfig(utils.GlobalObject.WorkerPool)
}

// NewWorkerPoolWithConfig 创建新的工作池
// 使用自定义配置
func NewWorkerPoolWithConfig(config utils.WorkerPoolConfig) *WorkerPool {
	return &WorkerPool{
		coreWorkers:       config.CoreWorkers,
		maxWorkers:        config.MaxWorkers,
		queueSize:         config.QueueSize,
		idleTimeout:       time.Duration(config.IdleTimeout) * time.Second,
		currentWorkers:    0,
		WorkerPool:        make(chan chan zinterface.IRequest, config.MaxWorkers),
		JobQueue:          make(chan zinterface.IRequest, config.QueueSize),
		workers:           make(map[uint32]*Worker),
		stopChan:          make(chan bool),
		isStopped:         false,
		nextWorkerID:      1,
		dispatcherStarted: false,
		timerStarted:      false,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.isStopped {
		return
	}

	// 启动任务调度器（只启动一次）
	if !wp.dispatcherStarted {
		wp.dispatcherStarted = true
		wp.wg.Add(1)
		go wp.dispatch()
	}

	// 启动核心工作线程（快速启动）
	for wp.currentWorkers < wp.coreWorkers {
		wp.createWorker()
	}

	// 启动空闲工作线程检测（只启动一次）
	if !wp.timerStarted {
		wp.timerStarted = true
		wp.wg.Add(1)
		go wp.checkIdleWorkers()
	}
}

// createWorker 创建一个工作线程
// 注意：调用此方法前必须持有wp.mutex锁
func (wp *WorkerPool) createWorker() {
	workerID := wp.nextWorkerID
	wp.nextWorkerID++

	// 判断是否是核心工作线程
	isCore := wp.currentWorkers < wp.coreWorkers
	worker := NewWorker(workerID, wp.WorkerPool, isCore)
	worker.Start()

	wp.workers[workerID] = worker
	wp.currentWorkers++

	utils.GlobalLogger.Info("Worker %d started (core: %t), current workers: %d", workerID, isCore, wp.currentWorkers)
}

// dispatch 任务调度器
// 负责将请求分配给工作线程
func (wp *WorkerPool) dispatch() {
	defer wp.wg.Done()

	for {
		select {
		// 接收请求
		case request := <-wp.JobQueue:
			wp.handleRequest(request)
		// 接收停止信号
		case <-wp.stopChan:
			wp.mutex.Lock()
			for _, worker := range wp.workers {
				worker.Stop()
			}
			wp.isStopped = true
			wp.mutex.Unlock()
			utils.GlobalLogger.Info("WorkerPool dispatcher stopped")
			return
		}
	}
}

// handleRequest 处理请求
// 实现工作线程的动态创建逻辑
func (wp *WorkerPool) handleRequest(request zinterface.IRequest) {
	// 尝试获取一个可用的工作线程
	select {
	// 尝试获取一个可用的工作线程
	case workerJobQueue := <-wp.WorkerPool:
		// 将请求分配给工作线程
		workerJobQueue <- request
	default:
		// 没有可用的工作线程，检查是否可以创建新的工作线程
		wp.mutex.RLock()
		canCreateNewWorker := wp.currentWorkers < wp.maxWorkers
		wp.mutex.RUnlock()

		if canCreateNewWorker {
			// 创建新的工作线程
			wp.mutex.Lock()
			wp.createWorker()
			wp.mutex.Unlock()
			// 等待新的工作线程注册到WorkerPool
			workerJobQueue := <-wp.WorkerPool
			workerJobQueue <- request
		} else {
			// 已经达到最大工作线程数，阻塞等待可用的工作线程
			workerJobQueue := <-wp.WorkerPool
			workerJobQueue <- request
		}
	}
}

// AddRequest 添加请求到工作池
func (wp *WorkerPool) AddRequest(request zinterface.IRequest) {
	// 如果工作池已经停止，拒绝请求
	wp.mutex.RLock()
	if wp.isStopped {
		wp.mutex.RUnlock()
		utils.GlobalLogger.Warn("WorkerPool is stopped, request rejected")
		return
	}
	wp.mutex.RUnlock()

	// 尝试添加请求到队列
	select {
	case wp.JobQueue <- request:
		// 请求添加成功
	default:
		// 队列已满，记录警告
		utils.GlobalLogger.Warn("WorkerPool job queue is full, request rejected")
	}
}

// checkIdleWorkers 检查并销毁空闲的非核心工作线程
func (wp *WorkerPool) checkIdleWorkers() {
	defer wp.wg.Done()

	ticker := time.NewTicker(wp.idleTimeout / 2) // 每一半的超时时间检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wp.checkIdleWorkersImpl()
		case <-wp.stopChan:
			return
		}
	}
}

// checkIdleWorkersImpl 检查并销毁空闲的非核心工作线程的实现
func (wp *WorkerPool) checkIdleWorkersImpl() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.isStopped || wp.currentWorkers <= wp.coreWorkers {
		return
	}

	now := time.Now()
	var workersToStop []uint32

	// 查找空闲的非核心工作线程
	for workerID, worker := range wp.workers {
		if worker.isCore {
			continue
		}

		// 检查是否超时
		if now.Sub(worker.lastActivity) > wp.idleTimeout {
			workersToStop = append(workersToStop, workerID)
		}
	}

	// 停止空闲的工作线程
	for _, workerID := range workersToStop {
		worker := wp.workers[workerID]
		worker.Stop()
		delete(wp.workers, workerID)
		wp.currentWorkers--
		utils.GlobalLogger.Info("Worker %d stopped due to idle timeout, current workers: %d", workerID, wp.currentWorkers)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	wp.mutex.Lock()
	if wp.isStopped {
		wp.mutex.Unlock()
		return
	}

	// 发送停止信号
	close(wp.stopChan)

	// 停止所有工作线程
	for _, worker := range wp.workers {
		worker.Stop()
	}

	// 清空请求队列
	close(wp.JobQueue)

	wp.isStopped = true
	wp.mutex.Unlock()

	// 等待所有goroutine结束
	wp.wg.Wait()

	utils.GlobalLogger.Info("WorkerPool stopped, total workers: %d", len(wp.workers))
}

// GetWorkerSize 获取当前工作线程数
func (wp *WorkerPool) GetWorkerSize() uint32 {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.currentWorkers
}

// GetCoreWorkers 获取核心工作线程数
func (wp *WorkerPool) GetCoreWorkers() uint32 {
	return wp.coreWorkers
}

// GetMaxWorkers 获取最大工作线程数
func (wp *WorkerPool) GetMaxWorkers() uint32 {
	return wp.maxWorkers
}

// GetQueueSize 获取请求队列大小
func (wp *WorkerPool) GetQueueSize() uint32 {
	return wp.queueSize
}

// GetIdleTimeout 获取空闲超时时间
func (wp *WorkerPool) GetIdleTimeout() time.Duration {
	return wp.idleTimeout
}

// IsStopped 检查工作池是否已停止
func (wp *WorkerPool) IsStopped() bool {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.isStopped
}
