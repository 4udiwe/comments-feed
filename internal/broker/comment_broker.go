package broker

import (
	"sync"
)

type Comment struct {
	ID      string
	Content string
	PostID  string
}

// CommentBroker — in-memory pub/sub механизм для доставки событий о новых комментариях.
//
// Используется для реализации GraphQL Subscriptions: клиенты подписываются на комментарии
// конкретного поста и получают обновления асинхронно через каналы.
//
// Особенности:
// - подписки хранятся в памяти процесса (не подходит для multi-instance без внешнего брокера)
// - публикация неблокирующая (медленные подписчики могут пропускать события)
// - требуется явная отписка для предотвращения утечек
//
// В production-окружении может быть заменён на внешний брокер (например, Redis или Kafka)
// без изменения сервисного слоя.
type CommentBroker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *Comment
}

func NewCommentBroker() *CommentBroker {
	return &CommentBroker{
		subscribers: make(map[string][]chan *Comment),
	}
}

func (b *CommentBroker) Subscribe(postID string) (<-chan *Comment, func()) {
	ch := make(chan *Comment, 1)

	b.mu.Lock()
	b.subscribers[postID] = append(b.subscribers[postID], ch)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs := b.subscribers[postID]
		for i, sub := range subs {
			if sub == ch {
				b.subscribers[postID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, unsubscribe
}

func (b *CommentBroker) Publish(postID string, comment *Comment) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers[postID] {
		select {
		case ch <- comment:
		default:
			// не блокируемся
		}
	}
}