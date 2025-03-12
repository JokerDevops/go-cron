package main
import "os"
import "os/exec"
import "strings"
import "sync"
import "os/signal"
import "syscall"
import "github.com/robfig/cron"
import "sync/atomic"

var activeTasks int32 // 用于记录当前活跃的任务数

func execute(command string, args []string)() {

    println("executing:", command, strings.Join(args, " "))

    cmd := exec.Command(command, args...)

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    // Run 已经阻塞 不需要在Wait
    if err := cmd.Run(); err != nil {
      println("Error executing command:", err)
    }
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
    if len(os.Args) < 3 {
      println("Usage: go-cron <schedule> <command> [args...]")
      os.Exit(1)
    }

    var schedule string = os.Args[1]
    var command string = os.Args[2]
    var args []string = os.Args[3:len(os.Args)]

    wg := &sync.WaitGroup{}

    c := cron.New()
    println("new cron:", schedule)

    c.AddFunc(schedule, func() {
        wg.Add(1)
        atomic.AddInt32(&activeTasks, 1)  // 增加活跃任务数
        defer wg.Done()
        defer atomic.AddInt32(&activeTasks, -1)  // 减少活跃任务数
        execute(command, args)
    })

    return c, wg
}

func start(c *cron.Cron, wg *sync.WaitGroup) {
    c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
    println("Stopping")
    c.Stop()
    if atomic.LoadInt32(&activeTasks) > 0 {
      fmt.Println("Waiting for running jobs to finish...")
      wg.Wait()
    } else {
      fmt.Println("No active jobs. Exiting immediately.")
    }
    fmt.Println("All jobs finished. Exiting.")
    os.Exit(0)
}

func main() {

    c, wg := create()

    go start(c, wg)

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    println(<-ch)

    stop(c, wg)
}


