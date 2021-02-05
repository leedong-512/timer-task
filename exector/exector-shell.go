package exector

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

/*type Job interface {
	Exector()
}*/
type ExectorShell struct {
	Command string
	Cmd *exec.Cmd
}
type ExectorResult struct {
	Code int
	Info string
}

func NewExectorShell(Command string) *ExectorShell {
	return &ExectorShell{
		Command: Command,
	}
}
func (e *ExectorShell) Execute() {
	var wg sync.WaitGroup
	// 使用上下文实现强制结束
	ctx, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	wg.Add(1)
	var (
		output []byte
		err    error
	)
	go func() {

		// 将context传入cmd实现channel的监听和任务强杀
		if runtime.GOOS == "windows" {
			cmdstr :=strings.Fields(e.Command)
			e.Cmd = exec.CommandContext(ctx, cmdstr[0], cmdstr[1])
		} else {
			e.Cmd = exec.CommandContext(ctx, "bash", "-c", e.Command)
		}
		if output, err = e.Cmd.CombinedOutput(); err != nil {
			log.Println(err)
		}
		fmt.Println(string(output))
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	// 强制结束命令
	wg.Wait()

	/*return &ExectorResult{
		Code: 200,
		Info: string(output),
	}*/
}