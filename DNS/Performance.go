package main

// 测 CPU 和 Mem 实现，不过 Jemeter PerfMon 也支持这个功能，所以该代码作为备选项。
import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func Test() {
	// 打开或创建一个日志文件
	file, err := os.OpenFile("stats.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for {
		// 获取 CPU 使用率（百分比）
		cpuPercent, err := cpu.Percent(time.Second, false)
		if err != nil {
			fmt.Fprintf(file, "获取CPU信息失败: %v\n", err)
			continue
		}

		// 获取内存信息
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			fmt.Fprintf(file, "获取内存信息失败: %v\n", err)
			continue
		}

		// 获取Go运行时内存信息（可选）
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// 写入文件
		log := fmt.Sprintf("[%s] CPU使用率: %.2f%%, 内存使用率: %.2f%%, Go内存: %.2fMB\n",
			time.Now().Format("2006-01-02 15:04:05"),
			cpuPercent[0],
			vmStat.UsedPercent,
			float64(memStats.Alloc)/1024/1024,
		)
		file.WriteString(log)

		// 每5秒记录一次
		time.Sleep(5 * time.Second)
	}
}
