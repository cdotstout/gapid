// Copyright (C) 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/gapid/core/assert"
	"github.com/google/gapid/core/event/task"
	"github.com/google/gapid/core/log"
)

var delay = time.Millisecond * 100

func waitForQueued(s *Scheduler, n int) {
	for s.NumTasksQueued() != n {
		time.Sleep(time.Millisecond)
	}
}

type testExecutor struct {
	got [][]int
	val interface{}
	err error
}

func (t *testExecutor) exec(ctx context.Context, l []Executable, b Batch) {
	tasks := make([]int, len(l))
	for i, e := range l {
		tasks[i] = e.Task.(int)
		e.Result(t.val, t.err)
	}
	sort.Ints(tasks)
	t.got = append(t.got, tasks)
}

func setup(t *testing.T) (context.Context, *testExecutor, *Scheduler, *sync.WaitGroup) {
	ctx := log.Testing(t)
	e := &testExecutor{val: 321}
	s := New(ctx, e.exec)
	wg := &sync.WaitGroup{}
	return ctx, e, s, wg
}

func TestSingle(t *testing.T) {
	ctx, e, s, _ := setup(t)
	val, err := s.Schedule(ctx, 123, Batch{})
	assert.To(t).For("val").That(val).Equals(321)
	assert.To(t).For("err").ThatError(err).Succeeded()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{[]int{123}})
}

func TestManyBatchedWithDuration(t *testing.T) {
	ctx, e, s, wg := setup(t)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, i, Batch{Precondition: delay})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	waitForQueued(s, 5)
	wg.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{[]int{0, 1, 2, 3, 4}})
}

func TestManyBatchedWithTime(t *testing.T) {
	ctx, e, s, wg := setup(t)
	precondition := time.Now().Add(delay)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, i, Batch{Precondition: precondition})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	waitForQueued(s, 5)
	wg.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{[]int{0, 1, 2, 3, 4}})
}

func TestManyBatched(t *testing.T) {
	ctx, e, s, wg := setup(t)
	precondition, fence := task.NewSignal()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, i, Batch{Precondition: precondition})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	waitForQueued(s, 5)
	fence(ctx)
	wg.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{[]int{0, 1, 2, 3, 4}})
}

func TestManySeparateKeys(t *testing.T) {
	ctx, e, s, wg := setup(t)
	precondition, fence := task.NewSignal()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, i, Batch{Precondition: precondition, Key: i})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	waitForQueued(s, 5)
	fence(ctx)
	wg.Wait()
	assert.To(t).For("got").ThatSlice(e.got).IsLength(5)
}

func TestManySeparatePreconditions(t *testing.T) {
	ctx, e, s, _ := setup(t)
	preconditionA, fenceA := task.NewSignal()
	preconditionB, fenceB := task.NewSignal()
	preconditionC, fenceC := task.NewSignal()
	wgA, wgB, wgC := sync.WaitGroup{}, sync.WaitGroup{}, sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wgA.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, 30+i, Batch{Precondition: preconditionA})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wgA.Done()
		}(i)
	}
	for i := 0; i < 2; i++ {
		wgC.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, 20+i, Batch{Precondition: preconditionC})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wgC.Done()
		}(i)
	}
	for i := 0; i < 1; i++ {
		wgB.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, 10+i, Batch{Precondition: preconditionB})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wgB.Done()
		}(i)
	}

	waitForQueued(s, 3+2+1)

	fenceA(ctx)
	wgA.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{
		[]int{30, 31, 32},
	})

	fenceB(ctx)
	wgB.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{
		[]int{30, 31, 32},
		[]int{10},
	})

	fenceC(ctx)
	wgC.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{
		[]int{30, 31, 32},
		[]int{10},
		[]int{20, 21},
	})
}

func TestManySeparatePriorities(t *testing.T) {
	ctx, e, s, wg := setup(t)
	precondition, fence := task.NewSignal()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, 30+i, Batch{Precondition: precondition, Priority: 3})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, 10+i, Batch{Precondition: precondition, Priority: 1})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			val, err := s.Schedule(ctx, 20+i, Batch{Precondition: precondition, Priority: 2})
			assert.To(t).For("val %v", i).That(val).Equals(321)
			assert.To(t).For("err %v", i).ThatError(err).Succeeded()
			wg.Done()
		}(i)
	}
	waitForQueued(s, 3+1+2)
	fence(ctx)
	wg.Wait()
	assert.To(t).For("got").ThatSlice(e.got).DeepEquals([][]int{
		[]int{30, 31, 32},
		[]int{20, 21},
		[]int{10},
	})
}

func TestCancel(t *testing.T) {
	ctx, e, s, wg := setup(t)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			ctx, cancel := task.WithCancel(ctx)
			if i&1 == 1 {
				cancel()
			}
			val, err := s.Schedule(ctx, i, Batch{})
			if err == nil {
				assert.To(t).For("val %v", i).That(val).Equals(321)
			} else {
				assert.To(t).For("val %v", i).That(val).Equals(nil)
			}
			assert.To(t).For("err %v", i).That(err).Equals(task.StopReason(ctx))
			wg.Done()
		}(i)
	}
	wg.Wait()
	sum := 0
	for _, l := range e.got {
		sum += len(l)
	}
	assert.To(t).For("sum").That(sum).Equals(3)
}
