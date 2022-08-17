//	Package workerpool for creating a queue for deleting links.
package workerpool

import (
	"context"
	"sync"
	"time"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type InputWorker struct {
	mu     sync.Mutex
	ch     chan interfaces.Task
	done   chan struct{}
	index  int
	ticker *time.Ticker
	ctx    context.Context
}

type OutputWorker struct {
	id   int
	ch   chan interfaces.Task
	done chan struct{}
	db   interfaces.Storage
	ctx  context.Context
}

func NewInputWorker(ch chan interfaces.Task, done chan struct{}, ctx context.Context) *InputWorker {
	index := 0
	ticker := time.NewTicker(10 * time.Second)
	return &InputWorker{
		ch:     ch,
		done:   done,
		index:  index,
		ticker: ticker,
		ctx:    ctx,
	}
}

func NewOutputWorker(id int, ch chan interfaces.Task, done chan struct{}, ctx context.Context, db interfaces.Storage) *OutputWorker {
	return &OutputWorker{
		id:   id,
		ch:   ch,
		done: done,
		ctx:  ctx,
		db:   db,
	}
}

func (w *InputWorker) Do(t interfaces.Task) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ch <- t
	w.index++
	if w.index == 20 {
		w.done <- struct{}{}
		w.index = 0
	}
}

func (w *InputWorker) Loop() error {
	for {
		select {
		case <-w.ctx.Done():
			w.ticker.Stop()
			return nil
		case <-w.ticker.C:
			w.done <- struct{}{}
			w.index = 0
		}
	}
}

func (w *OutputWorker) Do() error {
	models := make([]interfaces.Task, 0, 20)
	for {
		select {
		case <-w.ctx.Done():
			return nil
		case <-w.done:
			if len(w.ch) == 0 {
				break
			}
			for task := range w.ch {
				models = append(models, task)
				if len(w.ch) == 0 {
					if err := w.db.DelBatchShortURLs(models); err != nil {
						return err
					}
					models = nil
					break
				}
			}
		}
	}
}
