package entrez

import (
	"context"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/config"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/output"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/pkg/errors"
)

// EntrezService 定义了固定的 Entrez 操作管道
type EntrezService struct {
	Config   *config.Config // 配置参数
	executor *QueryExecutor // 查询执行器
}

// NewEntrezService 创建一个新的 EntrezService 实例
func NewEntrezService(config *config.Config) *EntrezService {
	// 创建 EntrezService 实例
	service := &EntrezService{
		Config:   config,
		executor: NewQueryExecutor(config),
	}

	return service
}

// ExecuteQueries 执行查询并返回结果通道
func (s *EntrezService) ExecuteQueries(ctx context.Context, queries []*types.Query) (<-chan *types.QueryResult, error) {
	return s.executor.executeQueries(ctx, queries)
}

// ProcessResults 处理查询结果
func (s *EntrezService) ProcessResults(ctx context.Context, results <-chan *types.QueryResult, outputFile string, resultWriter output.Writer) error {
	if resultWriter == nil {
		return customerrors.NewEmptyResultError("result writer is nil")
	}

	// 创建一个新的 context，避免外部 context 取消影响写入操作
	writeCtx, cancel := context.WithTimeout(ctx, s.Config.Runtime.WriteTimeout)
	defer cancel()

	// 只有当有结果需要处理时才初始化 writer
	if err := resultWriter.SetHeaders(nil); err != nil {
		return errors.Wrapf(customerrors.ErrSaveResult, "failed to set headers: %v", err)
	}

	done := make(chan struct{})
	var writeErr error

	// 启动一个 goroutine 处理结果写入
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				writeErr = errors.Wrapf(customerrors.ErrSaveResult, "panic during result processing: %v", r)
				logcdl.Error("recovered from panic in processresults: %v", r)
			}
		}()

		// 检查结果通道是否为空
		if results == nil {
			writeErr = customerrors.NewEmptyResultError("results channel is nil")
			return
		}

		// 写入结果流
		if err := resultWriter.WriteResultStream(writeCtx, results); err != nil {
			writeErr = err
			logcdl.Error("error writing result stream: %v", err)
			return
		}
	}()

	// 等待写入完成或上下文取消
	select {
	case <-done:
		if writeErr != nil {
			return writeErr
		}
		logcdl.Success("result processing completed")
	case <-writeCtx.Done():
		logcdl.Error("context cancelled before processing completed")
		return writeCtx.Err()
	}

	// 保存结果
	if err := resultWriter.Save(outputFile); err != nil {
		return errors.Wrapf(customerrors.ErrSaveResult, "failed to save results: %v", err)
	}

	return nil
}
