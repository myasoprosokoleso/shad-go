//go:build !solution

package tparallel

type T struct {
	hasParallel      chan struct{}
	startParallel    chan struct{}
	done             chan struct{}
	parallelSubTests []*T
}

func (t *T) Parallel() {
	close(t.hasParallel)
	<-t.startParallel
}

func (t *T) Run(subtest func(t *T)) {
	subT := processTest(subtest)

	select {
	case <-subT.hasParallel:
		t.parallelSubTests = append(t.parallelSubTests, subT)
	case <-subT.done:
	}
}

func Run(topTests []func(t *T)) {
	var allTs, parallelTs []*T

	for _, test := range topTests {
		t := processTest(test)
		allTs = append(allTs, t)

		select {
		case <-t.hasParallel:
			parallelTs = append(parallelTs, t)
		case <-t.done:
		}
	}

	for _, t := range parallelTs {
		close(t.startParallel)
	}
	for _, t := range allTs {
		<-t.done
	}
}

func processTest(test func(t *T)) *T {
	t := &T{
		hasParallel:   make(chan struct{}),
		startParallel: make(chan struct{}),
		done:          make(chan struct{}),
	}

	go func() {
		test(t)
		for _, p := range t.parallelSubTests {
			close(p.startParallel)
		}
		for _, p := range t.parallelSubTests {
			<-p.done
		}
		close(t.done)
	}()

	return t
}
