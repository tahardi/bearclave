package clock

type RDTSCFunc func() int64

type TSCTimer struct {
	start     int64
	stop      int64
	frequency int64
	rdtsc     RDTSCFunc
}

func NewTSCTimer() (*TSCTimer, error) {
	frequency, err := GetTSCFrequency()
	if err != nil {
		return nil, timerError("", nil)
	}
	return NewTSCTimerWithRDTSC(frequency, RDTSC)
}

func NewTSCTimerWithRDTSC(frequency int64, rdtsc RDTSCFunc) (*TSCTimer, error) {
	return &TSCTimer{
		frequency: frequency,
		rdtsc:     rdtsc,
	}, nil
}

func (t *TSCTimer) Start() {
	t.start = t.rdtsc()
}

func (t *TSCTimer) Stop() {
	t.stop = t.rdtsc()
}

func (t *TSCTimer) Reset() {
	t.start = 0
	t.stop = 0
}

func (t *TSCTimer) ElapsedNanoseconds() int64 {
	cycles := t.stop - t.start
	return (cycles * Billion) / t.frequency
}
