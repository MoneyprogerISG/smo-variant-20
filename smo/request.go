package smo

type Request struct {
	sourceID         int
	number           int
	genTime          float64
	bufferInTime     float64
	serviceStartTime float64
	serviceEndTime   float64
}

const notSpecified = -1

func NewRequest(sourceID, number int, genTime float64) *Request {
	return &Request{
		sourceID:         sourceID,
		number:           number,
		genTime:          genTime,
		bufferInTime:     notSpecified,
		serviceStartTime: notSpecified,
		serviceEndTime:   notSpecified,
	}
}

func (r *Request) SourceID() int {
	return r.sourceID
}

func (r *Request) Number() int {
	return r.number
}

func (r *Request) GenTime() float64 {
	return r.genTime
}

func (r *Request) SetBufferInTime(t float64) {
	r.bufferInTime = t
}

func (r *Request) SetServiceStartTime(t float64) {
	r.serviceStartTime = t
}

func (r *Request) SetServiceEndTime(t float64) {
	r.serviceEndTime = t
}
