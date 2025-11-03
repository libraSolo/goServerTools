package pubsub

// Middleware 泛型中间件类型
type Middleware[T any] func(subject string, content T, next Handler[T])

// PubSubWithMiddleware 是带有中间件功能的发布订阅服务
type PubSubWithMiddleware[T any] struct {
	*GenericPubSub[T]
	middlewares []Middleware[T]
}

// NewPubSubWithMiddleware 创建一个带中间件的发布订阅服务实例
func NewPubSubWithMiddleware[T any]() *PubSubWithMiddleware[T] {
	return &PubSubWithMiddleware[T]{
		GenericPubSub: NewGenericPubSub[T](),
		middlewares:   []Middleware[T]{},
	}
}

// Use 添加一个或多个中间件
func (ps *PubSubWithMiddleware[T]) Use(middlewares ...Middleware[T]) {
	ps.middlewares = append(ps.middlewares, middlewares...)
}

// Subscribe 订阅主题，并应用中间件
func (ps *PubSubWithMiddleware[T]) Subscribe(subscriberID string, subject string, handler Handler[T]) error {
	wrappedHandler := ps.wrapHandler(handler)
	return ps.GenericPubSub.Subscribe(subscriberID, subject, wrappedHandler)
}

// wrapHandler 将处理器包装在中间件链中
func (ps *PubSubWithMiddleware[T]) wrapHandler(handler Handler[T]) Handler[T] {
	if len(ps.middlewares) == 0 {
		return handler
	}

	wrapped := handler
	for i := len(ps.middlewares) - 1; i >= 0; i-- {
		mw := ps.middlewares[i]
		current := wrapped
		wrapped = func(subject string, content T) {
			mw(subject, content, current)
		}
	}
	return wrapped
}