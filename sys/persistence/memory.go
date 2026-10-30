// Package persistence is the persistence. The handlers only know the Repository
// interface; here it is a map in memory, it could be a database.
package persistence

import (
	"errors"
	"slices"
	"sync"
)

var ErrNotFound = errors.New("not found")

// Result is one page of a search.
type Result[T any] struct {
	Items []T
	Page  int // the page that was returned (it may differ from the one asked)
	Pages int
	Total int
}

type Repository[T any] interface {
	// Find returns the items that match, ordered by id, one page at a time. Pages start at 1.
	Find(match func(T) bool, page, size int) (Result[T], error)
	// All returns every item that matches, ordered by id.
	All(match func(T) bool) ([]T, error)
	FindOne(id int) (T, error)
	// Save creates the item if its id is 0, and replaces it otherwise.
	Save(item T) (T, error)
	Delete(id int) error
}

type Memory[T any] struct {
	mu     sync.Mutex
	items  map[int]T
	nextID int
	id     func(T) int
	withID func(T, int) T
}

func NewMemory[T any](id func(T) int, withID func(T, int) T, seed ...T) *Memory[T] {
	m := &Memory[T]{items: map[int]T{}, nextID: 1, id: id, withID: withID}
	for _, item := range seed {
		m.Save(item)
	}
	return m
}

func (m *Memory[T]) Find(match func(T) bool, page, size int) (Result[T], error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]int, 0, len(m.items))
	for id, item := range m.items {
		if match == nil || match(item) {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	pages := max(1, (len(ids)+size-1)/size)
	page = min(max(page, 1), pages)
	from, to := (page-1)*size, min(page*size, len(ids))
	items := make([]T, 0, to-from)
	for _, id := range ids[from:to] {
		items = append(items, m.items[id])
	}
	return Result[T]{Items: items, Page: page, Pages: pages, Total: len(ids)}, nil
}

func (m *Memory[T]) All(match func(T) bool) ([]T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]int, 0, len(m.items))
	for id, item := range m.items {
		if match == nil || match(item) {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	items := make([]T, 0, len(ids))
	for _, id := range ids {
		items = append(items, m.items[id])
	}
	return items, nil
}

func (m *Memory[T]) FindOne(id int) (T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return item, ErrNotFound
	}
	return item, nil
}

func (m *Memory[T]) Save(item T) (T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.id(item) == 0 {
		item = m.withID(item, m.nextID)
		m.nextID++
	} else if _, ok := m.items[m.id(item)]; !ok {
		return item, ErrNotFound
	}
	m.items[m.id(item)] = item
	return item, nil
}

func (m *Memory[T]) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}
