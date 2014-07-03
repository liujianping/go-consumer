/*

	type Consumer interface{
		Put(request interface{})
		Resume(context Context, fork int) error
		Close() error

		Running() bool
		Depth() int64
		Finished() int64
	}

	consumer.NewMemoryConsumer()
	consumer.NewPersistConsumer()

	consumer.Resume(context Context, fork int) error
	consumer.Put(request interface{})

	consumer.Pause() error
	consumer.UnPause() error

	consumer.Flush() error
	consumer.Clear() error
	consumer.Close() error

	consumer.Running() bool
	consumer.Depth() int
	
	type Context interface {
		Encode(request interface{}) []byte, error
		Decode(data []byte) interface{}, error
		Do(request interface{}) error
	}

*/


package consumer 
