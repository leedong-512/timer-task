package api

import (
	"fmt"
	"log"
	"net/http"
	"task/v1/config"
	"task/v1/models"
	"task/v1/scheduler"

	"github.com/gin-gonic/gin"
)

type Transport interface {
	HttpServer()
}

type HttpTransport struct {
	Engine   *gin.Engine
	Sched    *scheduler.Scheduler
	StopHttp chan bool
}

func NewHttpTransport(sched *scheduler.Scheduler) *HttpTransport {
	return &HttpTransport{
		Sched: sched,
	}
}

func (h *HttpTransport) HttpServer() *http.Server {
	h.Engine = gin.Default()
	h.ApiRoutes()
	srv := &http.Server{
		Addr:    config.HttpAddr,
		Handler: h.Engine,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	fmt.Println("Listen:", config.HttpAddr)
	//go h.Engine.Run(config.HttpAddr)
	return srv
}

func (h *HttpTransport) ApiRoutes() {
	h.Engine.GET("/ping", func(cxt *gin.Context) {
		cxt.JSON(http.StatusOK, gin.H{
			"code": 0,
			"info": "pong",
		})
	})

	v1 := h.Engine.Group("/v1")
	v1.POST("/jobs", h.jobAddHandler)
	v1.GET("/jobs", h.jobsGethandler)

	jobs := v1.Group("/jobs")
	jobs.GET("/:job", h.jobGetHandler)
	jobs.PUT("/:job", h.jobUpdateHandler)
	jobs.DELETE("/:job", h.jobDeleteHandler)
	jobs.GET("/:job/run", h.jobSingleRunHandler)
	jobs.GET("/:job/executions", h.executionsHandler)
}

// 添加任务
func (h *HttpTransport) jobAddHandler(ctx *gin.Context) {
	//var job models.Job
	job := models.Job{}
	if err := ctx.Bind(&job); err != nil {
		//ctx.Writer.WriteString(fmt.Sprintf("Unable to parse payload: %s.", err))
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		//log.Error(err)
		return
	}
	//ctx.Bind(&job)
	if err := job.Validate(); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
		//ctx.Writer.WriteString(fmt.Sprintf("Job contains invalid value: %s.", err))
		return
	}
	store := models.NewStore()
	if err := store.SetJob(&job); err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}
	err := h.Sched.AddJob(&job)
	if err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, responseData(SuccessCode, "添加成功", nil))
}

// 获取所有任务
func (h *HttpTransport) jobsGethandler(ctx *gin.Context) {
	jobs, err := h.Sched.GetJobs()
	if err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, responseData(SuccessCode, "请求成功", jobs))
}

// 获取单个任务
func (h *HttpTransport) jobGetHandler(ctx *gin.Context) {
	jobName := ctx.Param("job")

	job, err := h.Sched.GetJob(jobName)
	if err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, responseData(SuccessCode, "请求成功", job))
}

// 更新任务
func (h *HttpTransport) jobUpdateHandler(ctx *gin.Context) {
	var job models.Job

	ctx.Bind(&job)
	h.Sched.UpdateJob(&job)

	ctx.JSON(http.StatusOK, responseData(SuccessCode, "请求成功", job))
}

// 删除任务
func (h *HttpTransport) jobDeleteHandler(ctx *gin.Context) {
	jobName := ctx.Param("job")
	err := h.Sched.RemoveJob(jobName)
	if err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, responseData(SuccessCode, "删除成功", nil))
}

// 手动执行任务
func (h *HttpTransport) jobSingleRunHandler(ctx *gin.Context) {
	jobName := ctx.Param("job")
	err := h.Sched.SingleRunJob(jobName)
	if err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}
	ctx.JSON(http.StatusOK, responseData(SuccessCode, "执行成功", nil))
}

// 获取任务执行情况
func (h *HttpTransport) executionsHandler(ctx *gin.Context) {
	jobName := ctx.Param("job")
	exections, err := h.Sched.GetJobExecutions(jobName)
	if err != nil {
		ctx.JSON(http.StatusOK, responseData(UnknownErrorCode, err.Error(), nil))
		return
	}
	ctx.JSON(http.StatusOK, responseData(SuccessCode, "请求成功", exections))
}

func responseData(code int32, msg string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
