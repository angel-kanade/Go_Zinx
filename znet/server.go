package znet

// IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	Name      string
	IPVersion string
	IP        string
	Port      int
}

func (s *Server) Start() {

}

func (s *Server) Serve() {

}

func (s *Server) Stop() {

}
