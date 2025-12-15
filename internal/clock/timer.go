package clock

type Timer interface {
	Start()
	Stop()
	Reset()
	ElapsedNanoseconds() int64
}
