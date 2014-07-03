package consumer

type Context interface{
	Encode(request interface{}) ([]byte, error)
	Decode(data []byte) (interface{}, error)
	Do(request interface{}) error
}

type Consumer interface{
	Put(request interface{}) error
	Resume(context Context, fork int) error
	Close() error

	Running() bool
	Depth() int64
}

