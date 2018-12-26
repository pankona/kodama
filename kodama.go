package kodama

import (
	"context"
	"fmt"
	"net"
	"strconv"

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
	WorkerNum int
	RetryNum  int
	Validator Validator
	Worker    Worker
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
	service.RegisterKodamaServer(s, k)

	k.runDispatcher()

	return k.server.Serve(listen)
}

var (
	errOK      = &service.Error{ErrCode: service.ErrCode_OK}
	errBusy    = &service.Error{ErrCode: service.ErrCode_BUSY}
	errGeneric = &service.Error{ErrCode: service.ErrCode_GENERIC}
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

func (k *Server) runDispatcher() {
	workerCh := make(chan struct{}, k.WorkerNum)
	for {
		workerCh <- struct{}{}
		go func() {
			select {
			case desc := <-k.jobQueue:
			retry:
				for i := 0; i < k.RetryNum; i++ {
					if err := k.Worker.Work(desc); err == nil {
						break retry
					}
				}
				// TODO: retry count exceeded. notify to system admin

			}
			// TODO: notify result to caller

			<-workerCh
		}()
	}
}
