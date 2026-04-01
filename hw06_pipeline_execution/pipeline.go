package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	currentIn := wrapInputWithDone(in, done)
	for _, stage := range stages {
		currentIn = runStage(currentIn, done, stage)
	}
	return currentIn
}

func wrapInputWithDone(in In, done In) Out {
	return forwardUntilDone(done, in)
}

func runStage(in In, done In, stage Stage) Out {
	out := make(Bi)
	stageOut := stage(forwardUntilDone(done, in))

	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				consumeUntilClosed(stageOut)
				return
			case v, ok := <-stageOut:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					consumeUntilClosed(stageOut)
					return
				}
			}
		}
	}()
	return out
}

func forwardUntilDone(done In, in In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					return
				}
			}
		}
	}()

	return out
}

func consumeUntilClosed(in In) {
	for drain := range in {
		_ = drain // intentionally empty: drain channel to prevent goroutine leak on cancellation
	}
}
