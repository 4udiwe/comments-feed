package transactor

import "context"

//go:generate go tool mockgen -source=transactor.go -destination=../../internal/mocks/mock_transactor.go
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
}