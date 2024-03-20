package filter

import "slices"

type Chain struct {
	filterers []Filterer
}

func (c *Chain) Filter(paths []string) (filtered []string, err error) {
	filtered = slices.Clone(paths)
	for _, f := range c.filterers {
		if filtered, err = f.Filter(filtered); err != nil {
			return filtered, err
		}
	}
	return filtered, nil
}

func NewChain(filterers ...Filterer) *Chain {
	return &Chain{filterers: slices.Clone(filterers)}
}
