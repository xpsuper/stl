
package adapter
--------------------

Package adapter exposes functionality to convert an arbitrary map[string]interface{} into a native Go structure.

The Go structure can be arbitrarily complex, containing slices, other structs, etc. and the decoder will properly decode nested maps and so on into the proper structures in the native Go struct. See the examples to see what the decoder is capable of.

*   [func Decode(input interface{}, output interface{}) error](#Decode)
*   [func DecodeHookExec( raw DecodeHookFunc, from reflect.Type, to reflect.Type, data interface{}) (interface{}, error)](#DecodeHookExec)
*   [func DecodeMetadata(input interface{}, output interface{}, metadata *Metadata) error](#DecodeMetadata)
*   [func WeakDecode(input, output interface{}) error](#WeakDecode)
*   [func WeakDecodeMetadata(input interface{}, output interface{}, metadata *Metadata) error](#WeakDecodeMetadata)
*   [func WeaklyTypedHook( f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error)](#WeaklyTypedHook)
*   [type DecodeHookFunc](#DecodeHookFunc)

*   [func ComposeDecodeHookFunc(fs ...DecodeHookFunc) DecodeHookFunc](#ComposeDecodeHookFunc)
*   [func StringToIPHookFunc() DecodeHookFunc](#StringToIPHookFunc)
*   [func StringToIPNetHookFunc() DecodeHookFunc](#StringToIPNetHookFunc)
*   [func StringToSliceHookFunc(sep string) DecodeHookFunc](#StringToSliceHookFunc)
*   [func StringToTimeDurationHookFunc() DecodeHookFunc](#StringToTimeDurationHookFunc)
*   [func StringToTimeHookFunc(layout string) DecodeHookFunc](#StringToTimeHookFunc)

*   [type DecodeHookFuncKind](#DecodeHookFuncKind)
*   [type DecodeHookFuncType](#DecodeHookFuncType)
*   [type Decoder](#Decoder)

*   [func NewDecoder(config *DecoderConfig) (*Decoder, error)](#NewDecoder)
*   [func (d *Decoder) Decode(input interface{}) error](#Decoder.Decode)

*   [type DecoderConfig](#DecoderConfig)
*   [type Error](#Error)

*   [func (e *Error) Error() string](#Error.Error)
*   [func (e *Error) WrappedErrors() []error](#Error.WrappedErrors)

*   [type Metadata](#Metadata)

#### Examples [¶](#pkg-examples)

*   [Decode](#example-Decode)
*   [Decode (EmbeddedStruct)](#example-Decode--EmbeddedStruct)
*   [Decode (Errors)](#example-Decode--Errors)
*   [Decode (Metadata)](#example-Decode--Metadata)
*   [Decode (Tags)](#example-Decode--Tags)
*   [Decode (WeaklyTypedInput)](#example-Decode--WeaklyTypedInput)

```
func Decode(input interface{}, output interface{}) error

```

Decode takes an input structure and uses reflection to translate it to the output structure. output must be a pointer to a map or struct.

Code:

```
type Person struct {
    Name   string
    Age    int
    Emails []string
    Extra  map[string]string
}

// This input can come from anywhere, but typically comes from
// something like decoding JSON where we're not quite sure of the
// struct initially.
input := map[string]interface{}{
    "name":   "Mitchell",
    "age":    91,
    "emails": []string{"one", "two", "three"},
    "extra": map[string]string{
        "twitter": "mitchellh",
    },
}

var result Person
err := Decode(input, &result)
if err != nil {
    panic(err)
}

fmt.Printf("%#v", result)

```

Output:

```
adapter.Person{Name:"Mitchell", Age:91, Emails:[]string{"one", "two", "three"}, Extra:map[string]string{"twitter":"mitchellh"}}


```

Code:

```
// Squashing multiple embedded structs is allowed using the squash tag.
// This is demonstrated by creating a composite struct of multiple types
// and decoding into it. In this case, a person can carry with it both
// a Family and a Location, as well as their own FirstName.
type Family struct {
    LastName string
}
type Location struct {
    City string
}
type Person struct {
    Family    `adapter:",squash"`
    Location  `adapter:",squash"`
    FirstName string
}

input := map[string]interface{}{
    "FirstName": "Mitchell",
    "LastName":  "Hashimoto",
    "City":      "San Francisco",
}

var result Person
err := Decode(input, &result)
if err != nil {
    panic(err)
}

fmt.Printf("%s %s, %s", result.FirstName, result.LastName, result.City)

```

Output:

```
Mitchell Hashimoto, San Francisco


```

Code:

```
type Person struct {
    Name   string
    Age    int
    Emails []string
    Extra  map[string]string
}

// This input can come from anywhere, but typically comes from
// something like decoding JSON where we're not quite sure of the
// struct initially.
input := map[string]interface{}{
    "name":   123,
    "age":    "bad value",
    "emails": []int{1, 2, 3},
}

var result Person
err := Decode(input, &result)
if err == nil {
    panic("should have an error")
}

fmt.Println(err.Error())

```

Output:

```
5 error(s) decoding:

* 'Age' expected type 'int', got unconvertible type 'string'
* 'Emails[0]' expected type 'string', got unconvertible type 'int'
* 'Emails[1]' expected type 'string', got unconvertible type 'int'
* 'Emails[2]' expected type 'string', got unconvertible type 'int'
* 'Name' expected type 'string', got unconvertible type 'int'


```

Code:

```
type Person struct {
    Name   string
    Age    int
    Emails []string
}

// This input can come from anywhere, but typically comes from
// something like decoding JSON, generated by a weakly typed language
// such as PHP.
input := map[string]interface{}{
    "name":   123,                      // number => string
    "age":    "42",                     // string => number
    "emails": map[string]interface{}{}, // empty map => empty array
}

var result Person
config := &DecoderConfig{
    WeaklyTypedInput: true,
    Result:           &result,
}

decoder, err := NewDecoder(config)
if err != nil {
    panic(err)
}

err = decoder.Decode(input)
if err != nil {
    panic(err)
}

fmt.Printf("%#v", result)

```

Output:

```
adapter.Person{Name:"123", Age:42, Emails:[]string{}}


```

DecodeHookExec executes the given decode hook. This should be used since it'll naturally degrade to the older backwards compatible DecodeHookFunc that took reflect.Kind instead of reflect.Type.

```
func DecodeMetadata(input interface{}, output interface{}, metadata *Metadata) error

```

DecodeMetadata is the same as Decode, but is shorthand to enable metadata collection. See DecoderConfig for more info.

```
func WeakDecode(input, output interface{}) error

```

WeakDecode is the same as Decode but is shorthand to enable WeaklyTypedInput. See DecoderConfig for more info.

```
func WeakDecodeMetadata(input interface{}, output interface{}, metadata *Metadata) error

```

WeakDecodeMetadata is the same as Decode, but is shorthand to enable both WeaklyTypedInput and metadata collection. See DecoderConfig for more info.

WeaklyTypedHook is a DecodeHookFunc which adds support for weak typing to the decoder.

Note that this is significantly different from the WeaklyTypedInput option of the DecoderConfig.

```
type DecodeHookFunc interface{}

```

DecodeHookFunc is the callback function that can be used for data transformations. See "DecodeHook" in the DecoderConfig struct.

The type should be DecodeHookFuncType or DecodeHookFuncKind. Either is accepted. Types are a superset of Kinds (Types can return Kinds) and are generally a richer thing to use, but Kinds are simpler if you only need those.

The reason DecodeHookFunc is multi-typed is for backwards compatibility: we started with Kinds and then realized Types were the better solution, but have a promise to not break backwards compat so we now support both.

ComposeDecodeHookFunc creates a single DecodeHookFunc that automatically composes multiple DecodeHookFuncs.

The composed funcs are called in order, with the result of the previous transformation.

StringToIPHookFunc returns a DecodeHookFunc that converts strings to net.IP

StringToIPNetHookFunc returns a DecodeHookFunc that converts strings to net.IPNet

StringToSliceHookFunc returns a DecodeHookFunc that converts string to []string by splitting on the given sep.

StringToTimeDurationHookFunc returns a DecodeHookFunc that converts strings to time.Duration.

StringToTimeHookFunc returns a DecodeHookFunc that converts strings to time.Time.

DecodeHookFuncKind is a DecodeHookFunc which knows only the Kinds of the source and target types.

DecodeHookFuncType is a DecodeHookFunc which has complete information about the source and target types.

```
type Decoder struct {
    // contains filtered or unexported fields
}

```

A Decoder takes a raw interface value and turns it into structured data, keeping track of rich error information along the way in case anything goes wrong. Unlike the basic top-level Decode method, you can more finely control how the Decoder behaves using the DecoderConfig structure. The top-level Decode method is just a convenience that sets up the most basic Decoder.

NewDecoder returns a new decoder for the given configuration. Once a decoder has been returned, the same configuration must not be used again.

Decode decodes the given raw interface to the target pointer specified by the configuration.

```
type DecoderConfig struct {
    // DecodeHook, if set, will be called before any decoding and any
    // type conversion (if WeaklyTypedInput is on). This lets you modify
    // the values before they're set down onto the resulting struct.
    //
    // If an error is returned, the entire decode will fail with that
    // error.
    DecodeHook DecodeHookFunc

    // If ErrorUnused is true, then it is an error for there to exist
    // keys in the original map that were unused in the decoding process
    // (extra keys).
    ErrorUnused bool

    // ZeroFields, if set to true, will zero fields before writing them.
    // For example, a map will be emptied before decoded values are put in
    // it. If this is false, a map will be merged.
    ZeroFields bool

    // If WeaklyTypedInput is true, the decoder will make the following
    // "weak" conversions:
    //
    //   - bools to string (true = "1", false = "0")
    //   - numbers to string (base 10)
    //   - bools to int/uint (true = 1, false = 0)
    //   - strings to int/uint (base implied by prefix)
    //   - int to bool (true if value != 0)
    //   - string to bool (accepts: 1, t, T, TRUE, true, True, 0, f, F,
    //     FALSE, false, False. Anything else is an error)
    //   - empty array = empty map and vice versa
    //   - negative numbers to overflowed uint values (base 10)
    //   - slice of maps to a merged map
    //   - single values are converted to slices if required. Each
    //     element is weakly decoded. For example: "4" can become []int{4}
    //     if the target type is an int slice.
    //
    WeaklyTypedInput bool

    // Metadata is the struct that will contain extra metadata about
    // the decoding. If this is nil, then no metadata will be tracked.
    Metadata *Metadata

    // Result is a pointer to the struct that will contain the decoded
    // value.
    Result interface{}

    // The tag name that adapter reads for field names. This
    // defaults to "adapter"
    TagName string
}

```

DecoderConfig is the configuration that is used to create a new decoder and allows customization of various aspects of decoding.

```
type Error struct {
    Errors []string
}

```

Error implements the error interface and can represents multiple errors that occur in the course of a single decode.

WrappedErrors implements the errwrap.Wrapper interface to make this return value more useful with the errwrap and go-multierror libraries.

```
type Metadata struct {
    // Keys are the keys of the structure which were successfully decoded
    Keys []string

    // Unused is a slice of keys that were found in the raw value but
    // weren't decoded since there was no matching field in the result interface
    Unused []string
}

```

Metadata contains information about decoding a structure that is tedious or difficult to get otherwise.