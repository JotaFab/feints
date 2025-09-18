package player

import (
    "container/heap"
    "sync"
)


// A songHeap is a min-heap of songs.
type songHeap []*Song

func (h songHeap) Len() int           { return len(h) }
func (h songHeap) Less(i, j int) bool { return h[i].index < h[j].index }
func (h songHeap) Swap(i, j int) {
    h[i], h[j] = h[j], h[i]
    h[i].index = i
    h[j].index = j
}

// Push adds an item to the heap.
func (h *songHeap) Push(x any) {
    n := len(*h)
    item := x.(*Song)
    item.index = n
    *h = append(*h, item)
}

// Pop removes and returns the smallest item from the heap.
func (h *songHeap) Pop() any {
    old := *h
    n := len(old)
    item := old[n-1]
    old[n-1] = nil // Avoid memory leaks
    item.index = -1
    *h = old[0 : n-1]
    return item
}

// Remove removes an item from the heap by its index.
func (h *songHeap) Remove(i int) any {
    return heap.Remove(h, i)
}

// List returns a copy of the songs in the heap.
func (h songHeap) List() []Song {
    list := make([]Song, h.Len())
    for i, song := range h {
        list[i] = *song
    }
    return list
}

// NewsongHeap creates and initializes an empty songHeap.
func NewsongHeap() *songHeap {
    h := &songHeap{}
    heap.Init(h)
    return h
}

// SongQueue is the song queue.
type SongQueue struct {
    sync.RWMutex
    items songHeap
}

// NewSongQueue creates a new song queue instance.
func NewSongQueue() *SongQueue {
    q := &SongQueue{
        items: *NewsongHeap(),
    }
    return q
}

// Push adds a song to the queue.
func (q *SongQueue) Push(song Song) {
    q.Lock()
    defer q.Unlock()
    heap.Push(&q.items, &song)
}

// Pop removes and returns the song with the highest priority (the next one).
func (q *SongQueue) Pop() *Song {
    q.Lock()
    defer q.Unlock()
    if q.items.Len() == 0 {
        return nil
    }
    item := heap.Pop(&q.items).(*Song)
    return item
}

// List returns all songs in the queue.
func (q *SongQueue) List() []Song {
    q.RLock()
    defer q.RUnlock()
    // It is safer to return a copy
    return q.items.List()
}

// Clear empties the queue.
func (q *SongQueue) Clear() {
    q.Lock()
    defer q.Unlock()
    q.items = *NewsongHeap()
}

