package rest

type PollAsync chan struct{}

func NewPollAsync() PollAsync {
	return PollAsync(make(chan struct{}, 1))
}

func (pa PollAsync) Stop() {
	pa <- struct{}{}
}
