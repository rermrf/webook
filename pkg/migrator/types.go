package migrator

type Entity interface {
	ID() int64
	Equal(dst Entity) bool
}
