package linq

// Reverse 反转集合
func (q LinQuery) Reverse() LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()

			items := []interface{}{}
			for item, ok := next(); ok; item, ok = next() {
				items = append(items, item)
			}

			index := len(items) - 1
			return func() (item interface{}, ok bool) {
				if index < 0 {
					return
				}

				item, ok = items[index], true
				index--
				return
			}
		},
	}
}
