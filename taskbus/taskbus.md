# taskbus

taskbus is a utility library for manipulating synchronous and taskbushronous flows.

### Install

```bash
$ go get github.com/xpsuper/stl/taskbus
```
### All
```go
taskbus.All(tasks taskier) (Results, error)
```

All will execute all the functions Allly using goroutines.

- `tasks` is an internal interface which accept two taskbus public types:
    - `Tasks` is a list of functions to be executed Allly.
    - `MapTasks` is a map of string of functions to be executed Allly.

The type of `Results` value depends on the type of tasks passed to the function. See [type Results](#type-results)

All errors ocurred in the functions will be returned. See [returning error](#returning-error) and [type Errors](#type-errors).

### Max
```go
taskbus.Max(tasks taskier) (Results, error)
```

Max will execute all the functions in Max. It creates multiple goroutines and distributes the functions execution among them.
The number of goroutines  defaults to `GOMAXPROCS`. If the number of active goroutines is equal to `GOMAXPROCS` and there're more functions to execute, these functions will wait until one of functions being executed finishes its job.

- `tasks` is an internal interface which accept two taskbus public types: 
    - `taskbus.Tasks` is a list of functions to be executed in Max.
    - `taskbus.MapTasks` is a map of string of functions to be executed in Max.

The type of `Results` value depends on the type of tasks passed to the function. See [type Results](#type-results).

All errors ocurred in the functions will be returned. See [returning error](#returning-error) and [type Errors](#type-errors).

### Chain
```go
taskbus.Chain(tasks Tasks, args ...interface{}) ([]interface[}, error)
```

Chain will execute all the functions in sequence, each returning their results to the next. If the last returning value of the function is of type `error`, then this value will not be passed to the next function. 

- `tasks` is a list of functions that will be executed in series.
- `args` are optional parameters that will be passed to the first task.

Chain returns the results of the last task as `[]interface{}` and `error`. 

If an error occur in any of the functions to be executed, the next function will not be executed, and the error will be returned to the caller. See [returning error](#returning-error).

### <a name="type-results"></a>Type Results

The `Results` is the type that is returned by `Max` and `All`:

```go
type Results interface {
      Index(int) []interface{}  // Gets values by index
      Key(string) []interface{} // Gets values by key
      Len() int                 // Gets the length of the results
      Keys() []string           // Gets the keys of the results
}
```

The underlying type of `Results` will be different depending on the type of tasks passed to either `Max` or `All`.
There are two taskbus public types that can be passed to both functions:

 - `Tasks` is a list of functions that will be executed.
 - `MapTasks` is a map of string of functions to be executed.

When using `Tasks`, the underlying `Results` will be `[][]interface{}`:
```go
res, err : = taskbus.Max(taskbus.Tasks{
        func1,
        func2,
        ...funcN,
})
```
We can get the results of the the second function:

```go
res.Index(1)
```

Or we can iterate over the results using the  `Len()` method:

```go
for i := 0; i < res.Len(); i++ {
    fmt.Println(res.Index(i))
}
```

When using `MapTasks`, the underlying `Results` will be `map[string][]interface{}`:

```go
res, err : = taskbus.All(taskbus.MapTasks{
        "one"  : funcOne,
        "two"  : funcTwo,
        "three": funcThree,
})
```

We can get the results of function "three":

```go
res.Key("three")
```
Or we can also iterate over the map of results by using the `Keys()` method:

```go
for k := range res.Keys() {
    fmt.Println(res.Key(k))
}
```

### <a name="type-errors"></a>Type Errors

If errors occur in any function executed by `All` or `Max` an instance of `Errors` will be returned.
`Errors` implements the `error` interface, so in order to test if an error occurred, check if the returned error is not nil,
if it's not type cast it to `Errors`:

```go
_, err : = taskbus.Max(taskbus.Tasks{func1, func2, ...funcN})

if err != nil {
        MaxErrors := err.(taskbus.Errors)

        for _, e := range MaxErrors {
                fmt.Println(e.Error())
        }
}
```

### <a name="returning-error"></a>Returning error

In order for taskbus to identify if an error occured, the error **must** be the last returning value of the function:

```go
_, err := taskbus.Chain(taskbus.Tasks{
        func () (int, error) {
                return 1, nil
        },
        // Function with error
        func (i int) (string, error) {
                if i > 0 {
                    // This line will interrupt the execution flow
                    return "", errors.New("Error occurred")
                }
                return "Ok", nil
        },
        // This function will not be executed.
        func (s string) {
            return
        }
});

if err != nil {
      fmt.Println(err.Error()); // "Error occurred"
}
```

# Examples 

### Chain

```go
import (
        "fmt"

        "github.com/xpsuper/stl/taskbus"
)

func fib(p, c int) (int, int) {
  return c, p + c
}

func main() {

        // execution in series.
        res, e := taskbus.Chain(taskbus.Tasks{
                fib,
                fib,
                fib,
                func(p, c int) (int, error) {
                        return c, nil
                },
        }, 0, 1)

        if e != nil {
              fmt.Printf("Error executing a Chain (%s)\n", e.Error())
        }

        fmt.Println(res[0].(int)) // Prints 3
}

```

### Max

```go
import (
        "fmt"
        "time"

        "github.com/xpsuper/stl/taskbus"
)

func main() {

        res, e := taskbus.Max(taskbus.MapTasks{
                "one": func() int {
                        for i := 'a'; i < 'a'+26; i++ {
                                fmt.Printf("%c ", i)
                        }
                        
                        return 1
                },
                "two": func() int {
                        time.Sleep(2 * time.Microsecond)
                        for i := 0; i < 27; i++ {
                                fmt.Printf("%d ", i)
                        }
                        
                        return 2
                },
                "three": func() int {
                        for i := 'z'; i >= 'a'; i-- {
                                fmt.Printf("%c ", i)
                        }
                        
                        return 3
                },
        })

        if e != nil {
                fmt.Printf("Errors [%s]\n", e.Error())
        }
        
        fmt.Println("Results from task 'two': %v", res.Key("two"))
}
```

### All

```go
import (
        "errors"
        "fmt"

        "github.com/xpsuper/stl/taskbus"
)

func main() {

        res, e := taskbus.All(taskbus.Tasks{
                func() int {
                        for i := 'a'; i < 'a'+26; i++ {
                                fmt.Printf("%c ", i)
                        }
                        return 0
                },
                func() error {
                        time.Sleep(3 * time.Microsecond)
                        for i := 0; i < 27; i++ {
                                fmt.Printf("%d ", i)
                        }
                        return errors.New("Error executing concurently")
                },
        })

        if e != nil {
                fmt.Printf("Errors [%s]\n", e.Error()) // output errors separated by space
        }

        fmt.Println("Result from function 0: %v", res.Index(0))
}
```
