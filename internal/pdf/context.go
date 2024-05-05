package pdf

type ContextStack struct {
	q      []any
	ofsset int
	size   int
}

type Tupple func(int, int, int)

func (ctx *ContextStack) Push(item any) {
	if ctx.q == nil {
		ctx.q = make([]any, 128)
		ctx.size = 0
		ctx.ofsset = -1
	}
	if ctx.ofsset > cap(ctx.q) {
		nq := make([]any, cap(ctx.q)<<1)
		copy(nq, ctx.q)
		ctx.q = nq
	}
	ctx.ofsset++
	ctx.q[ctx.ofsset] = item
	ctx.size++
}

func (ctx *ContextStack) PushD(a any, b any) {
	ctx.Push(b)
	ctx.Push(a)
}

func (ctx *ContextStack) PushT(a, b, c any) {
	ctx.Push(c)
	ctx.Push(b)
	ctx.Push(a)
}

func (ctx *ContextStack) PopD() (a any, b any) {
	return ctx.Pop(), ctx.Pop()
}

func (ctx *ContextStack) PopT() (any, any, any) {
	return ctx.Pop(), ctx.Pop(), ctx.Pop()
}

func (ctx *ContextStack) Peek() any {
	if ctx.size == 0 {
		panic("empty stack")
	}
	return ctx.q[ctx.ofsset]
}

func (ctx *ContextStack) Pop() (r any) {
	if ctx.size == 0 {
		panic("empty stack")
	}
	r = ctx.q[ctx.ofsset]
	ctx.ofsset--
	ctx.size--
	return
}

func PopSolo[T any](c *ContextStack) T {
	return c.Pop().(T)
}

func PopDual[T any](c *ContextStack) (T, T) {
	return c.Pop().(T), c.Pop().(T)
}

func PopTrio[T any](c *ContextStack) (T, T, T) {
	return c.Pop().(T), c.Pop().(T), c.Pop().(T)
}

func PopDuald2[T any, U any](c *ContextStack) (T, U) {
	return c.Pop().(T), c.Pop().(U)
}

func PopDuald3[T any, U any, V any](c *ContextStack) (T, U, V) {
	return c.Pop().(T), c.Pop().(U), c.Pop().(V)
}
