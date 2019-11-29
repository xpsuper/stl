### helper.Contains

Returns true if an element is present in a iteratee (slice, map,
string).

One frustrating thing in Go is to implement `contains` methods for each
type, for example:

``` {.sourceCode .go}
func ContainsInt(s []int, e int) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
```

this can be replaced by `helper.Contains`:

``` {.sourceCode .go}
// slice of string
helper.Contains([]string{"foo", "bar"}, "bar") // true

// slice of Foo ptr
helper.Contains([]*Foo{f}, f) // true
helper.Contains([]*Foo{f}, nil) // false

b := &Foo{
    ID:        2,
    FirstName: "Florent",
    LastName:  "Messa",
    Age:       28,
}

helper.Contains([]*Foo{f}, b) // false

// string
helper.Contains("florent", "rent") // true
helper.Contains("florent", "foo") // false

// even map
helper.Contains(map[int]string{1: "Florent"}, 1) // true
```

### helper.IndexOf

Gets the index at which the first occurrence of a value is found in an
array or return -1 if the value cannot be found.

``` {.sourceCode .go}
// slice of string
helper.IndexOf([]string{"foo", "bar"}, "bar") // 1
helper.IndexOf([]string{"foo", "bar"}, "gilles") // -1
```

### helper.LastIndexOf

Gets the index at which the last occurrence of a value is found in an
array or return -1 if the value cannot be found.

``` {.sourceCode .go}
// slice of string
helper.LastIndexOf([]string{"foo", "bar", "bar"}, "bar") // 2
helper.LastIndexOf([]string{"foo", "bar"}, "gilles") // -1
```

### helper.ToMap

Transforms a slice of structs to a map based on a `pivot` field.

``` {.sourceCode .go}
f := &Foo{
    ID:        1,
    FirstName: "Gilles",
    LastName:  "Fabio",
    Age:       70,
}

b := &Foo{
    ID:        2,
    FirstName: "Florent",
    LastName:  "Messa",
    Age:       80,
}

results := []*Foo{f, b}

mapping := helper.ToMap(results, "ID") // map[int]*Foo{1: f, 2: b}
```

### helper.Filter

Filters a slice based on a predicate.

``` {.sourceCode .go}
r := helper.Filter([]int{1, 2, 3, 4}, func(x int) bool {
    return x%2 == 0
}) // []int{2, 4}
```

### helper.Find

Finds an element in a slice based on a predicate.

``` {.sourceCode .go}
r := helper.Find([]int{1, 2, 3, 4}, func(x int) bool {
    return x%2 == 0
}) // 2
```

### helper.Map

Manipulates an iteratee (map, slice) and transforms it to another type:

-   map -\> slice
-   map -\> map
-   slice -\> map
-   slice -\> slice

``` {.sourceCode .go}
r := helper.Map([]int{1, 2, 3, 4}, func(x int) int {
    return x * 2
}) // []int{2, 4, 6, 8}

r := helper.Map([]int{1, 2, 3, 4}, func(x int) string {
    return "Hello"
}) // []string{"Hello", "Hello", "Hello", "Hello"}

r = helper.Map([]int{1, 2, 3, 4}, func(x int) (int, int) {
    return x, x
}) // map[int]int{1: 1, 2: 2, 3: 3, 4: 4}

mapping := map[int]string{
    1: "Florent",
    2: "Gilles",
}

r = helper.Map(mapping, func(k int, v string) int {
    return k
}) // []int{1, 2}

r = helper.Map(mapping, func(k int, v string) (string, string) {
    return fmt.Sprintf("%d", k), v
}) // map[string]string{"1": "Florent", "2": "Gilles"}
```

### helper.Get

Retrieves the value at path of struct(s).

``` {.sourceCode .go}
var bar *Bar = &Bar{
    Name: "Test",
    Bars: []*Bar{
        &Bar{
            Name: "Level1-1",
            Bar: &Bar{
                Name: "Level2-1",
            },
        },
        &Bar{
            Name: "Level1-2",
            Bar: &Bar{
                Name: "Level2-2",
            },
        },
    },
}

var foo *Foo = &Foo{
    ID:        1,
    FirstName: "Dark",
    LastName:  "Vador",
    Age:       30,
    Bar:       bar,
    Bars: []*Bar{
        bar,
        bar,
    },
}

helper.Get([]*Foo{foo}, "Bar.Bars.Bar.Name") // []string{"Level2-1", "Level2-2"}
helper.Get(foo, "Bar.Bars.Bar.Name") // []string{"Level2-1", "Level2-2"}
helper.Get(foo, "Bar.Name") // Test
```

`helper.Get` also handles `nil` values:

``` {.sourceCode .go}
bar := &Bar{
    Name: "Test",
}

foo1 := &Foo{
    ID:        1,
    FirstName: "Dark",
    LastName:  "Vador",
    Age:       30,
    Bar:       bar,
}

foo2 := &Foo{
    ID:        1,
    FirstName: "Dark",
    LastName:  "Vador",
    Age:       30,
} // foo2.Bar is nil

helper.Get([]*Foo{foo1, foo2}, "Bar.Name") // []string{"Test"}
helper.Get(foo2, "Bar.Name") // nil
```

### helper.Keys

Creates an array of the own enumerable map keys or struct field names.

``` {.sourceCode .go}
helper.Keys(map[string]int{"one": 1, "two": 2}) // []string{"one", "two"} (iteration order is not guaranteed)

foo := &Foo{
    ID:        1,
    FirstName: "Dark",
    LastName:  "Vador",
    Age:       30,
}

helper.Keys(foo) // []string{"ID", "FirstName", "LastName", "Age"} (iteration order is not guaranteed)
```

### helper.Values

Creates an array of the own enumerable map values or struct field
values.

``` {.sourceCode .go}
helper.Values(map[string]int{"one": 1, "two": 2}) // []string{1, 2} (iteration order is not guaranteed)

foo := &Foo{
    ID:        1,
    FirstName: "Dark",
    LastName:  "Vador",
    Age:       30,
}

helper.Values(foo) // []interface{}{1, "Dark", "Vador", 30} (iteration order is not guaranteed)
```

### helper.ForEach

Range over an iteratee (map, slice).

``` {.sourceCode .go}
helper.ForEach([]int{1, 2, 3, 4}, func(x int) {
    fmt.Println(x)
})
```

helper.ForEachRight ............

Range over an iteratee (map, slice) from the right.

``` {.sourceCode .go}
results := []int{}

helper.ForEachRight([]int{1, 2, 3, 4}, func(x int) {
    results = append(results, x)
})

fmt.Println(results) // []int{4, 3, 2, 1}
```

### helper.Chunk

Creates an array of elements split into groups with the length of the
size. If array can't be split evenly, the final chunk will be the
remaining element.

``` {.sourceCode .go}
helper.Chunk([]int{1, 2, 3, 4, 5}, 2) // [][]int{[]int{1, 2}, []int{3, 4}, []int{5}}
```

### helper.FlattenDeep

Recursively flattens an array.

``` {.sourceCode .go}
helper.FlattenDeep([][]int{[]int{1, 2}, []int{3, 4}}) // []int{1, 2, 3, 4}
```

### helper.Uniq

Creates an array with unique values.

``` {.sourceCode .go}
helper.Uniq([]int{0, 1, 1, 2, 3, 0, 0, 12}) // []int{0, 1, 2, 3, 12}
```

### helper.Drop

Creates an array/slice with n elements dropped from the beginning.

``` {.sourceCode .go}
helper.Drop([]int{0, 0, 0, 0}, 3) // []int{0}
```

### helper.Initial

Gets all but the last element of array.

``` {.sourceCode .go}
helper.Initial([]int{0, 1, 2, 3, 4}) // []int{0, 1, 2, 3}
```

### helper.Tail

Gets all but the first element of array.

``` {.sourceCode .go}
helper.Tail([]int{0, 1, 2, 3, 4}) // []int{1, 2, 3, 4}
```

### helper.Shuffle

Creates an array of shuffled values.

``` {.sourceCode .go}
helper.Shuffle([]int{0, 1, 2, 3, 4}) // []int{2, 1, 3, 4, 0}
```

### helper.Sum

Computes the sum of the values in an array.

``` {.sourceCode .go}
helper.Sum([]int{0, 1, 2, 3, 4}) // 10.0
helper.Sum([]interface{}{0.5, 1, 2, 3, 4}) // 10.5
```

### helper.Reverse

Transforms an array such that the first element will become the last,
the second element will become the second to last, etc.

``` {.sourceCode .go}
helper.Reverse([]int{0, 1, 2, 3, 4}) // []int{4, 3, 2, 1, 0}
```

### helper.SliceOf

Returns a slice based on an element.

``` {.sourceCode .go}
helper.SliceOf(f) // will return a []*Foo{f}
```

### helper.RandomInt

Generates a random int, based on a min and max values.

``` {.sourceCode .go}
helper.RandomInt(0, 100) // will be between 0 and 100
```

### helper.RandomString

Generates a random string with a fixed length.

``` {.sourceCode .go}
helper.RandomString(4) // will be a string of 4 random characters
```

### helper.Shard

Generates a sharded string with a fixed length and depth.

``` {.sourceCode .go}
helper.Shard("e89d66bdfdd4dd26b682cc77e23a86eb", 1, 2, false) // []string{"e", "8", "e89d66bdfdd4dd26b682cc77e23a86eb"}

helper.Shard("e89d66bdfdd4dd26b682cc77e23a86eb", 2, 2, false) // []string{"e8", "9d", "e89d66bdfdd4dd26b682cc77e23a86eb"}

helper.Shard("e89d66bdfdd4dd26b682cc77e23a86eb", 2, 2, true) // []string{"e8", "9d", "66", "bdfdd4dd26b682cc77e23a86eb"}
```
