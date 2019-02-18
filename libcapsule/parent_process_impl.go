package libcapsule

import "os"

type LinuxParentProcess struct {
}

func (p *LinuxParentProcess) pid() int {
	panic("implement me")
}

func (p *LinuxParentProcess) start() error {
	panic("implement me")
}

func (p *LinuxParentProcess) terminate() error {
	panic("implement me")
}

func (p *LinuxParentProcess) wait() (*os.ProcessState, error) {
	panic("implement me")
}

func (p *LinuxParentProcess) startTime() (uint64, error) {
	panic("implement me")
}

func (p *LinuxParentProcess) signal(os.Signal) error {
	panic("implement me")
}

func (p *LinuxParentProcess) externalDescriptors() []string {
	panic("implement me")
}

func (p *LinuxParentProcess) setExternalDescriptors(fds []string) {
	panic("implement me")
}
