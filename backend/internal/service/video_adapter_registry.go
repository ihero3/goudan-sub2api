package service

import "errors"

var (
	// ErrVideoAdapterRegistryNil indicates Resolve was called on an unset registry.
	ErrVideoAdapterRegistryNil = errors.New("video adapter registry is nil")
	// ErrVideoAdapterNotFound indicates no explicit or fallback adapter matched.
	ErrVideoAdapterNotFound = errors.New("video adapter not found")
)

// VideoAdapterMatcher lets a video adapter declare the account platform and/or
// model family it serves. Adapters may implement this in addition to
// VideoAdapter to opt into explicit routing; otherwise they are fallback
// adapters.
type VideoAdapterMatcher interface {
	Supports(platform, model string) bool
}

// VideoAdapterRegistry resolves an account/platform/model to a video adapter.
//
// Resolution is ordered: the first non-fallback adapter that reports Supports
// wins. At most one fallback adapter is allowed. If no explicit adapter
// matches, the fallback is returned. The fallback is how existing
// OpenAI-compatible accounts keep working after more providers are added.
type VideoAdapterRegistry struct {
	adapters []registryVideoAdapter
	fallback VideoAdapter
}

type registryVideoAdapter struct {
	adapter VideoAdapter
	matcher VideoAdapterMatcher
}

// NewVideoAdapterRegistry assembles an ordered adapter chain. Any adapter that
// implements VideoAdapterMatcher is treated as explicit; the final adapter is
// treated as the fallback. A nil registry is considered invalid and will panic
// on Resolve, catching wiring mistakes at startup.
func NewVideoAdapterRegistry(adapters ...VideoAdapter) *VideoAdapterRegistry {
	reg := &VideoAdapterRegistry{}
	if len(adapters) > 0 {
		reg.fallback = adapters[len(adapters)-1]
		adapters = adapters[:len(adapters)-1]
	}
	for _, a := range adapters {
		if a == nil {
			continue
		}
		entry := registryVideoAdapter{adapter: a}
		if matcher, ok := a.(VideoAdapterMatcher); ok {
			entry.matcher = matcher
		}
		reg.adapters = append(reg.adapters, entry)
	}
	return reg
}

// Resolve returns the adapter for the given platform and model.
func (r *VideoAdapterRegistry) Resolve(platform, model string) (VideoAdapter, error) {
	if r == nil {
		return nil, ErrVideoAdapterRegistryNil
	}
	for _, entry := range r.adapters {
		if entry.matcher != nil && entry.matcher.Supports(platform, model) {
			return entry.adapter, nil
		}
	}
	if r.fallback != nil {
		return r.fallback, nil
	}
	return nil, ErrVideoAdapterNotFound
}

// HasExplicitAdapter reports whether a non-fallback adapter matched the pair.
// It is useful for observability and for choosing whether a fallback path is
// being used.
func (r *VideoAdapterRegistry) HasExplicitAdapter(platform, model string) bool {
	if r == nil {
		return false
	}
	for _, entry := range r.adapters {
		if entry.matcher != nil && entry.matcher.Supports(platform, model) {
			return true
		}
	}
	return false
}
