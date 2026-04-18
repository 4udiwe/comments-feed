package transactor

import "context"

// NoOpTransactor реализует интерфейс Transactor без каких-либо реальных операций
// Используется в режиме in-memory хранилища, где транзакции не требуются
type NoOpTransactor struct{}

// NewNoOpTransactor создает новый no-op транзактор
func NewNoOpTransactor() *NoOpTransactor {
	return &NoOpTransactor{}
}

// WithinTransaction выполняет функцию без создания реальной транзакции
func (n *NoOpTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
