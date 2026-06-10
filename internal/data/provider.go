package data

import "github.com/waytofree-yang/right-side-trading/internal/domain"

type Provider interface {
	Load() (domain.DataSet, error)
}
