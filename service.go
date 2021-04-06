package gate

import (
	"fmt"
	"github.com/andyzhou/gate/proto"
	"github.com/andyzhou/gate/rpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

/*
 * gate rpc service, base on `grpc`
 *
 * - for service side
 */


//rpc service info
type Service struct {
	address string `rpc service address`
	rpc *rpc.Service `rpc service instance`
	service *grpc.Server `g-rpc server`
}

//construct
func NewService(port int) *Service {
	//self init
	address := fmt.Sprintf(":%d", port)
	this := &Service{
		address:address,
		rpc:rpc.NewService(),
		service:nil,
	}
	//create rpc service
	this.createService()
	return this
}

//stop
func (r *Service) Stop() {
	if r.service != nil {
		r.service.Stop()
	}
	if r.rpc != nil {
		r.rpc.Quit()
	}
}

/////////////////
//private func
/////////////////

//create rpc service
func (r *Service) createService() {
	var (
		tips string
		err error
	)

	//try listen tcp port
	listen, err := net.Listen("tcp", r.address)
	if err != nil {
		tips = "Create rpc service failed, error:" + err.Error()
		log.Println(tips)
		panic(tips)
	}

	//create rpc server with rpc stat support
	r.service = grpc.NewServer(
					grpc.StatsHandler(rpc.NewStat()),
				)

	//register call back
	gate.RegisterGateServiceServer(r.service, r.rpc)

	//begin rpc service
	go r.beginService(listen)
}

//begin rpc service
func (r *Service) beginService(listen net.Listener) {
	//service listen
	err := r.service.Serve(listen)
	if err != nil {
		tips := "Failed for rpc service, error:" + err.Error()
		log.Println(tips)
		panic(tips)
	}
}
