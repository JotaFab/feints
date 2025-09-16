package player

import "sync"


// SongQueue es una cola segura para concurrencia (FIFO)
type SongQueue struct {
	songs []Song
	mu    sync.Mutex
}

// NewSongQueue crea una nueva cola vacía
func NewSongQueue() *SongQueue {
	return &SongQueue{
		songs: []Song{},
	}
}

// Push agrega una canción al final de la cola
func (q *SongQueue) Push(s Song) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = append(q.songs, s)
}

// Pop devuelve y elimina la primera canción de la cola
func (q *SongQueue) Pop() *Song {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.songs) == 0 {
		return nil
	}
	s := q.songs[0]
	q.songs = q.songs[1:]
	return &s
}

// Peek devuelve la primera canción sin eliminarla
func (q *SongQueue) Peek() *Song {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.songs) == 0 {
		return nil
	}
	return &q.songs[0]
}

// Len devuelve la cantidad de canciones en la cola
func (q *SongQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.songs)
}

// List devuelve una copia de las canciones en la cola
func (q *SongQueue) List() []Song {
	q.mu.Lock()
	defer q.mu.Unlock()
	copied := make([]Song, len(q.songs))
	copy(copied, q.songs)
	return copied
}

// Clear elimina todas las canciones de la cola
func (q *SongQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = []Song{}
}
