package core

import "log"

// SongQueue es una cola FIFO controlada por un goroutine
type SongQueue struct {
    addCh   chan Song
    popCh   chan chan *Song
    listCh  chan chan []Song
    clearCh chan struct{}
}

// NewSongQueue crea e inicia una nueva cola FIFO
func NewSongQueue() *SongQueue {
    q := &SongQueue{
        addCh:   make(chan Song),
        popCh:   make(chan chan *Song),
        listCh:  make(chan chan []Song),
        clearCh: make(chan struct{}),
    }
    go q.loop()
    return q
}

// loop administra la cola de manera concurrente
func (q *SongQueue) loop() {
    var songs []Song
    for {
        select {
        case s := <-q.addCh:
            songs = append(songs, s)

        case reply := <-q.popCh:
            if len(songs) > 0 {
                song := songs[0]
                songs = songs[1:]
                reply <- &song
            } else {
                reply <- nil
            }

        case reply := <-q.listCh:
            copyList := make([]Song, len(songs))
            copy(copyList, songs)
            reply <- copyList

        case <-q.clearCh:
            songs = nil
        }
    }
}


// Pop obtiene y elimina la siguiente canción en FIFO
func (q *SongQueue) Pop() *Song {
    reply := make(chan *Song)
    q.popCh <- reply
    return <-reply
}

func (q *SongQueue) Push(song Song) {
    log.Printf("Push: %s", song.Title)
    q.addCh <- song
}

func (q *SongQueue) Clear() {
    log.Println("Queue cleared")
    q.clearCh <- struct{}{}
}

func (q *SongQueue) List() []Song {
    reply := make(chan []Song)
    q.listCh <- reply
    songs := <-reply
    log.Printf("List returned %d songs", len(songs))
    return songs
}
