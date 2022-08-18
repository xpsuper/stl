**Example 1: 查找2015年后生产的汽车的所有车主**

```go
import . "github.com/ahmetb/go-linq/v3"

type Car struct {
    year int
    owner, model string
}

...


var owners []string

From(cars).Where(func(c interface{}) bool {
	return c.(Car).year >= 2015
}).Select(func(c interface{}) interface{} {
	return c.(Car).owner
}).ToSlice(&owners)
```

也可以调用泛型方法, 如 `WhereT` and `SelectT` 来简化编码:

```go
var owners []string

From(cars).WhereT(func(c Car) bool {
	return c.year >= 2015
}).SelectT(func(c Car) string {
	return c.owner
}).ToSlice(&owners)
```

**Example 2: 找到写了最多书的作者**

```go
import . "github.com/ahmetb/go-linq/v3"

type Book struct {
	id      int
	title   string
	authors []string
}

author := From(books).SelectMany( // make a flat array of authors
	func(book interface{}) Query {
		return From(book.(Book).authors)
	}).GroupBy( // group by author
	func(author interface{}) interface{} {
		return author // author as key
	}, func(author interface{}) interface{} {
		return author // author as value
	}).OrderByDescending( // sort groups by its length
	func(group interface{}) interface{} {
		return len(group.(Group).Group)
	}).Select( // get authors out of groups
	func(group interface{}) interface{} {
		return group.(Group).Key
	}).First() // take the first author
```

**Example 3: 实现一个自定义方法，只保留大于指定阈值的值**

```go
type MyQuery Query

func (q MyQuery) GreaterThan(threshold int) Query {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if item.(int) > threshold {
						return
					}
				}

				return
			}
		},
	}
}

result := MyQuery(Range(1,10)).GreaterThan(5).Results()
```

**Example 4: 使用 “MapReduce” 在一个字符串中列出了使用通用函数的前5个最常用单词**

```go
var results []string

From(sentences).
	// split sentences to words
	SelectManyT(func(sentence string) Query {
		return From(strings.Split(sentence, " "))
	}).
	// group the words
	GroupByT(
		func(word string) string { return word },
		func(word string) string { return word },
	).
	// order by count
	OrderByDescendingT(func(wordGroup Group) int {
		return len(wordGroup.Group)
	}).
	// order by the word
	ThenByT(func(wordGroup Group) string {
		return wordGroup.Key.(string)
	}).
	Take(5).  // take the top 5
	// project the words using the index as rank
	SelectIndexedT(func(index int, wordGroup Group) string {
		return fmt.Sprintf("Rank: #%d, Word: %s, Counts: %d", index+1, wordGroup.Key, len(wordGroup.Group))
	}).
	ToSlice(&results)
```

