package task

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task/v1/api"
	"task/v1/models"
	"task/v1/scheduler"
	"time"
)

var task *Task

type Task struct {
	sched *scheduler.Scheduler
	store *models.Store
}

func init() {
	task = &Task{}
}

func Start() {
	// 调度器
	sched := scheduler.NewScheduler()
	task.sched = sched

	// 启动http服务
	httpTransport := api.NewHttpTransport(sched)
	httpServer := httpTransport.HttpServer()

	// 启动GRPC服务
	//l, err := net.Listen(config.NetWork, config.HttpAddr)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//srv := grpc.NewServer()
	//proto.RegisterTaskServer(srv, server.GrpcTaskServer{})
	//go srv.Serve(l)

	// 初始化将所有任务加入调度执行
	store := models.NewStore()
	jobs, err := store.GetJobs()
	if err != nil {
		log.Fatal(err)
	}
	sched.Start(jobs)
	task.store = store

	// 捕捉信号
	handleSignals(httpServer)
}

func handleSignals(httpServer *http.Server) {
	signCh := make(chan os.Signal, 2)
	signal.Notify(signCh, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)
	exitCh := make(chan int)

	go func() {

		select {
		case s := <-signCh:
			stopHttpServer(httpServer)
			fmt.Println("捕捉到信号:", s)
			// TODO 捕捉到信号后，此处需要做优雅退出处理，断开所有服务.比如切断http服务，数据库服务，调度等等

			go func() {
				if err := stop(); err != nil {
					fmt.Println(err)
					return
				}
			}()
			for {
				if models.GetRunningJobs() == 0 {
					fmt.Println("所有任务都处理完成了")
					break
				}
			}
			goto ExitProcess
		}

	ExitProcess:
		fmt.Println("Exit Service")
		exitCh <- 0
	}()

	code := <-exitCh
	os.Exit(code)
}

func stopHttpServer(httpServer *http.Server) {
	fmt.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("HttpServer Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 2 seconds.")
	}
	log.Println("Server exiting")
}

func stop() error {
	task.sched.Stop()
	if err := task.store.Close(); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
