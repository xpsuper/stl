package linq

// Union 返回两个集合的合集去重
func (q LinQuery) Union(q2 LinQuery) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			next2 := q2.Iterate()

			set := make(map[interface{}]bool)
			use1 := true

			return func() (item interface{}, ok bool) {
				if use1 {
					for item, ok = next(); ok; item, ok = next() {
						if _, has := set[item]; !has {
							set[item] = true
							return
						}
					}

					use1 = false
				}

				for item, ok = next2(); ok; item, ok = next2() {
					if _, has := set[item]; !has {
						set[item] = true
						return
					}
				}

				return
			}
		},
	}
}
