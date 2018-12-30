package kodama

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/pankona/kodama/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server   *grpc.Server
	jobQueue chan string
	*Configuration
}

type Configuration struct {
	Port      int
	QueueLen  int
	Validator Validator
}

func NewServer(cfg *Configuration) *Server {
	return &Server{Configuration: cfg}
}

func (k *Server) Run() error {
	p := ":" + strconv.Itoa(k.Port)
	listen, err := net.Listen("tcp", p)
	if err != nil {
		return fmt.Errorf("gRPC server failed to listen [%s]: %v", p, err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	k.server = s
	service.RegisterJobQueueServer(s, k)

	return k.server.Serve(listen)
}

// service.Errors
var (
	errOK      = &service.Error{ErrCode: service.ErrCode_OK}
	errBusy    = &service.Error{ErrCode: service.ErrCode_BUSY}
	errGeneric = &service.Error{ErrCode: service.ErrCode_GENERIC}
)

// ordinal errors
var (
	errNoJob = errors.New("job queue is empty")
)

func (k *Server) Push(ctx context.Context, job *service.Job) (*service.Error, error) {
	desc := job.Description
	if err := k.Validator.Validate(desc); err != nil {
		return errGeneric, fmt.Errorf("invalid job description: %v", err)
	}

	select {
	case k.jobQueue <- desc:
	default:
		return errBusy, fmt.Errorf("job queue is full.")
	}

	return errOK, nil
}

func (k *Server) Pop(context.Context, *service.Empty) (*service.Job, error) {
	select {
	case j := <-k.jobQueue:
		return &service.Job{Description: j}, nil
	case <-time.After(10 * time.Second):
		return nil, errNoJob
	}
}
