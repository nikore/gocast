package config

type Monitor interface {
	Execute() error
}

type portMonitor struct {
	protocol string
	port     int
}

func (p *portMonitor) Execute() error {

}

type execMonitor struct {
	cmd string
}

func (e *execMonitor) Execute() error {

}
