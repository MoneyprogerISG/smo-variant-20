package smo

type Statistics struct {
	generated        []int
	refused          []int
	served           []int
	totalWaitTime    []float64
	totalServiceTime []float64
	deviceBusyTime   []float64
	waitSqSum        []float64
	serviceSqSum     []float64
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
		waitSqSum:        make([]float64, sources),
		serviceSqSum:     make([]float64, sources),
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
	s.waitSqSum[r.sourceID] += waitTime * waitTime
}

func (s *Statistics) RecordServed(r *Request, now float64) {
	service := r.serviceEndTime - r.serviceStartTime
	s.served[r.sourceID]++
	s.totalServiceTime[r.sourceID] += service
	s.serviceSqSum[r.sourceID] += service * service
}

func (s *Statistics) AverageWait(sourceID int) float64 {
	if s.served[sourceID] == 0 {
		return 0
	}
	return s.totalWaitTime[sourceID] / float64(s.served[sourceID])
}

func (s *Statistics) AverageService(sourceID int) float64 {
	if s.served[sourceID] == 0 {
		return 0
	}
	return s.totalServiceTime[sourceID] / float64(s.served[sourceID])
}

func (s *Statistics) AverageSojourn(sourceID int) float64 {
	return s.AverageWait(sourceID) + s.AverageService(sourceID)
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

func (s *Statistics) WaitVariance(sourceID int) float64 {
	n := s.served[sourceID]
	if n == 0 {
		return 0
	}
	mean := s.totalWaitTime[sourceID] / float64(n)
	d := s.waitSqSum[sourceID]/float64(n) - mean*mean
	if d < 0 {
		return 0
	}
	return d
}

func (s *Statistics) ServiceVariance(sourceID int) float64 {
	n := s.served[sourceID]
	if n == 0 {
		return 0
	}
	mean := s.totalServiceTime[sourceID] / float64(n)
	d := s.serviceSqSum[sourceID]/float64(n) - mean*mean
	if d < 0 {
		return 0
	}
	return d
}

type SourceRow struct {
	SourceID           int
	Generated          int
	Refused            int
	RefusalProbability float64
	AvgSojourn         float64
	AvgWait            float64
	AvgService         float64
	WaitVariance       float64
	ServiceVariance    float64
}

type DeviceRow struct {
	DeviceID    int
	Utilization float64
}

type SummaryTable struct {
	TotalTime float64
	Sources   []SourceRow
	Devices   []DeviceRow
}

func (t SummaryTable) TotalRefusalProbability() float64 {
	generated, refused := 0, 0
	for _, row := range t.Sources {
		generated += row.Generated
		refused += row.Refused
	}
	if generated == 0 {
		return 0
	}
	return float64(refused) / float64(generated)
}

func (s *Statistics) BuildSummaryTable(totalTime float64) SummaryTable {
	sources := make([]SourceRow, len(s.generated))
	for i := range sources {
		sources[i] = SourceRow{
			SourceID:           i,
			Generated:          s.generated[i],
			Refused:            s.refused[i],
			RefusalProbability: s.RefusalProbability(i),
			AvgSojourn:         s.AverageSojourn(i),
			AvgWait:            s.AverageWait(i),
			AvgService:         s.AverageService(i),
			WaitVariance:       s.WaitVariance(i),
			ServiceVariance:    s.ServiceVariance(i),
		}
	}

	devices := make([]DeviceRow, len(s.deviceBusyTime))
	for j := range devices {
		devices[j] = DeviceRow{
			DeviceID:    j,
			Utilization: s.DeviceUtilization(j, totalTime),
		}
	}

	return SummaryTable{
		TotalTime: totalTime,
		Sources:   sources,
		Devices:   devices,
	}
}
