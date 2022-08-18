package linq

// DefaultIfEmpty 如果集合中的元素为空则置为默认值
func (q LinQuery) DefaultIfEmpty(defaultValue interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			state := 1

			return func() (item interface{}, ok bool) {
				switch state {
				case 1:
					item, ok = next()
					if ok {
						state = 2
					} else {
						item = defaultValue
						ok = true
						state = -1
					}
					return
				case 2:
					for item, ok = next(); ok; item, ok = next() {
						return
					}
					return
				}
				return
			}
		},
	}
}
