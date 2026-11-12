package main

import (
	"path/filepath"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"s3client/internal/awss3"
)

// P3：四路并发 + 每文件进度事件。
const maxConcurrent = 4
const transferEvent = "transfer:progress"

type TransferItemDTO struct {
	ID        string // 前端生成的唯一任务 id，用于把进度事件对应到进度行
	Name      string // 显示名（下载时可能是相对路径）
	Key       string // 对象 key（下载用）
	LocalPath string // 本地完整路径（上传用）
}

type TransferProgressDTO struct {
	ID      string
	Name    string
	Done    int64
	Total   int64
	Percent int
	Status  string // running / done / error
	Error   string
}

type TransferResultDTO struct {
	Total   int
	Success int
	Failed  int
	Errors  []string
}

// parallelRun 以 maxConcurrent 限流执行 n 个任务，返回与下标对应的错误切片。
// 抽成纯函数便于在不依赖 Wails runtime 的情况下测试限流行为。
func parallelRun(n, limit int, run func(i int) error) []error {
	errs := make([]error, n)
	var wg sync.WaitGroup
	sem := make(chan struct{}, limit)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			errs[idx] = run(idx)
		}(i)
	}
	wg.Wait()
	return errs
}

// emitProgress 推送单个任务的进度；回调高频触发时在 runTransfers 内做节流。
func (a *App) emitProgress(it TransferItemDTO, done, total int64, status, errMsg string) {
	if a.ctx == nil {
		return
	}
	pct := 0
	if total > 0 {
		pct = int(float64(done) / float64(total) * 100)
	}
	if status == "done" {
		pct = 100
	}
	wruntime.EventsEmit(a.ctx, transferEvent, TransferProgressDTO{
		ID:      it.ID,
		Name:    it.Name,
		Done:    done,
		Total:   total,
		Percent: pct,
		Status:  status,
		Error:   errMsg,
	})
}

// runTransfers 四路并发执行传输，逐任务推送 running / done / error 进度事件。
func (a *App) runTransfers(items []TransferItemDTO, run func(TransferItemDTO, func(int64, int64)) error) TransferResultDTO {
	result := TransferResultDTO{Total: len(items)}
	errs := parallelRun(len(items), maxConcurrent, func(i int) error {
		it := items[i]
		a.emitProgress(it, 0, 0, "running", "")

		var lastPct = -1
		var lastEmit time.Time
		progress := func(done, total int64) {
			pct := 0
			if total > 0 {
				pct = int(float64(done) / float64(total) * 100)
			}
			now := time.Now()
			// 节流：百分比变化至少 1%，或距上次推送超过 100ms
			if pct == lastPct && now.Sub(lastEmit) < 100*time.Millisecond {
				return
			}
			lastPct, lastEmit = pct, now
			a.emitProgress(it, done, total, "running", "")
		}

		if err := run(it, progress); err != nil {
			a.emitProgress(it, 0, 0, "error", err.Error())
			return err
		}
		a.emitProgress(it, 1, 1, "done", "")
		return nil
	})
	for _, e := range errs {
		if e != nil {
			result.Failed++
			result.Errors = append(result.Errors, e.Error())
		} else {
			result.Success++
		}
	}
	return result
}

// UploadFiles 并行上传多个本地文件到 bucket/prefix 目录下（key = prefix + 文件名）。
func (a *App) UploadFiles(acc AccountDTO, bucket, prefix string, items []TransferItemDTO) (TransferResultDTO, error) {
	m := toAccount(acc)
	return a.runTransfers(items, func(it TransferItemDTO, progress func(int64, int64)) error {
		return awss3.UploadFile(m, bucket, prefix+it.Name, filepath.FromSlash(it.LocalPath), progress)
	}), nil
}

// DownloadFiles 并行下载多个对象到 dir 目录（保留相对子目录结构）。
func (a *App) DownloadFiles(acc AccountDTO, bucket, dir string, items []TransferItemDTO) (TransferResultDTO, error) {
	m := toAccount(acc)
	return a.runTransfers(items, func(it TransferItemDTO, progress func(int64, int64)) error {
		local := filepath.Join(filepath.FromSlash(dir), filepath.FromSlash(it.Name))
		return awss3.DownloadFile(m, bucket, it.Key, local, progress)
	}), nil
}
