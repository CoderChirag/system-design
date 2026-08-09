package token_bucket

type TokenBucket interface {
	Allow() bool
}
