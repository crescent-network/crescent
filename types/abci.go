package types

import "context"

type ContextKey string

const (
	MidBlockContextKey = ContextKey("mid_block")
)

func MidBlockContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, MidBlockContextKey, struct{}{})
}

func IsMidBlockContext(ctx context.Context) bool {
	return ctx.Value(MidBlockContextKey) != nil
}
