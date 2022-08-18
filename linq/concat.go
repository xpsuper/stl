package linq

// Append 将元素插入集合的末尾，使其成为最后一个元素
func (q LinQuery) Append(item interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			appended := false

			return func() (interface{}, bool) {
				i, ok := next()
				if ok {
					return i, ok
				}

				if !appended {
					appended = true
					return item, true
				}

				return nil, false
			}
		},
	}
}

// Concat 连接两个集合
// Concat方法不同于Union方法，Concat 会返回两个结合中的所有原始元素, 而 Union 仅返回唯一元素
func (q LinQuery) Concat(q2 LinQuery) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			next2 := q2.Iterate()
			use1 := true

			return func() (item interface{}, ok bool) {
				if use1 {
					item, ok = next()
					if ok {
						return
					}

					use1 = false
				}

				return next2()
			}
		},
	}
}

// Prepend 将元素插入集合的头部，使其成为第一个元素
func (q LinQuery) Prepend(item interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			prepended := false

			return func() (interface{}, bool) {
				if prepended {
					return next()
				}

				prepended = true
				return item, true
			}
		},
	}
}
