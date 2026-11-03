package smo

type Statistics struct {
	generated        []int
	refused          []int
	served           []int
	totalWaitTime    []float64
	totalServiceTime []float64
	deviceBusyTime   []float64
}

func NewStatistics(sources, devices int) *Statistics {
	if sources <= 0 {
		panic("statistics: sources count must be positive")
	}
	if devices <= 0 {
		panic("statistics: devices count must be positive")
	}
	return &Statistics{
		generated:        make([]int, sources),
		refused:          make([]int, sources),
		served:           make([]int, sources),
		totalWaitTime:    make([]float64, sources),
		totalServiceTime: make([]float64, sources),
		deviceBusyTime:   make([]float64, devices),
	}
}

func (s *Statistics) RecordGenerated(r *Request) {
	s.generated[r.sourceID]++
}

func (s *Statistics) RecordRefused(r *Request, now float64) {
	s.refused[r.sourceID]++
}

func (s *Statistics) RecordWait(r *Request, waitTime float64) {
	s.totalWaitTime[r.sourceID] += waitTime
}

func (s *Statistics) RecordServed(r *Request, now float64) {
	s.served[r.sourceID]++
	s.totalServiceTime[r.sourceID] += r.serviceEndTime - r.serviceStartTime
}

func (s *Statistics) RecordBusyTime(d *Device, time float64) {
	s.deviceBusyTime[d.ID()] += time
}

func (s *Statistics) Generated(sourceID int) int {
	return s.generated[sourceID]
}

func (s *Statistics) Refused(sourceID int) int {
	return s.refused[sourceID]
}

func (s *Statistics) Served(sourceID int) int {
	return s.served[sourceID]
}

func (s *Statistics) RefusalProbability(sourceID int) float64 {
	if s.generated[sourceID] == 0 {
		return 0
	}
	return float64(s.refused[sourceID]) / float64(s.generated[sourceID])
}

func (s *Statistics) DeviceUtilization(deviceID int, totalTime float64) float64 {
	if totalTime <= 0 {
		return 0
	}
	return s.deviceBusyTime[deviceID] / totalTime
}
