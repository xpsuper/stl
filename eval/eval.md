#### 支持
*   Modifiers: `+` `-` `/` `*` `&` `|` `^` `**` `%` `>>` `<<`
*   Comparators: `>` `>=` `<` `<=` `==` `!=` `=~` `!~`
*   Logical ops: `||` `&&`
*   Numeric constants, as 64-bit floating point (`12345.678`)
*   String constants (double quotes: `"foobar"`)
*   Date function 'Date(x)', using any permutation of RFC3339, ISO8601, ruby date, or unix date
*   Boolean constants: `true` `false`
*   Parentheses to control order of evaluation `(` `)`
*   Json Arrays : `[1, 2, "foo"]`
*   Json Objects : `{"a":1, "b":2, "c":"foo"}`
*   Prefixes: `!` `-` `~`
*   Ternary conditional: `?` `:`
*   Null coalescence: `??`

#### Examples

```go
/******************************************************************/
// 基本用法
v, _ := stl.Eval("score > 65", map[string]interface{}{"score": 70})
logger.Debugf("Eval = %v", v) 
// OutPut: Eval = true

/******************************************************************/
// 字典和数组
v, _ = stl.Eval("foo.bar > 0", map[string]interface{}{
    "foo": map[string]interface{}{"bar": -1.},
})
logger.Debugf("Eval = %v", v)
// OutPut: Eval = false

v, _ := stl.Eval("foo["hey"] * bar[1]", map[string]interface{}{
    "foo": map[string]interface{}{"hey": 5.0},
    "bar": []interface{}{4, 5, 6},
})
logger.Debugf("Eval = %v", v)
// OutPut: Eval = 25.0

/******************************************************************/
// 定义函数
fun := stl.EvalFunction("strLen", func(args ...interface{}) (interface{}, error) {
    if len(args) != 1 {
        return nil, fmt.Errorf("strLen() takes exactly one argument")
    }
    if s, ok := args[0].(string); ok {
        return float64(len(s)), nil
    }
    return nil, fmt.Errorf("strLen() takes a string argument")
})
v, _ := stl.Eval(`strLen("someReallyLongInputString") <= 16`, nil, fun)
logger.Debugf("Eval = %v", v)
// OutPut: Eval = false

/******************************************************************/
// 对象属性和方法
type exampleType struct {
    Hello string
}
func (e exampleType) World() string {
    return "world"
}

v, _ := stl.Eval(`foo.Bar.Hello + foo.Bar.World()`, map[string]interface{}{
    "foo": struct{ Bar exampleType }{
        Bar: exampleType{Hello: "hello "},
    },
})
logger.Debugf("Eval = %v", v)
// OutPut: Eval = hello world

/******************************************************************/
// 预定义常量
eval, err := stl.EvalFull(stl.EvalConstant("max", 52)).NewEvaluable("value <= max")
if err != nil {
    logger.Errorf("EvalFull error: %v", err)
	return
}
for i := 50; i < 55; i++ {
    v, err := eval(context.Background(), map[string]interface{}{"value": i})
    if err != nil {
        logger.Errorf("Eval error: %v", err)
    }
    logger.Debugf("Eval = %v", v)
}
// OutPut:
// Eval = true
// Eval = true
// Eval = true
// Eval = false
// Eval = false

/******************************************************************/
// EvalBool
eval, err := stl.EvalFull().NewEvaluable("1 == x")
if err != nil {
    logger.Errorf("EvalFull error: %v", err)
	return
}
boolValue, err := eval.EvalBool(context.Background(), map[string]interface{}{"x": 1})
if err != nil {
    logger.Errorf("Eval error: %v", err)
}
if boolValue {
    logger.Debugf("Eval = true")
} else {
    logger.Debugf("Eval = false")
}
// OutPut: Eval = true

/******************************************************************/
// EvalFloat64
eval, err := stl.EvalFull().NewEvaluable("1 + x")
if err != nil {
    logger.Errorf("EvalFull error: %v", err)
    return
}
floatValue, err := eval.EvalFloat64(context.Background(), map[string]interface{}{"x": 5})
if err != nil {
    logger.Errorf("Eval error: %v", err)
}
logger.Debugf("Eval = %v", floatValue)
// OutPut: Eval = 6

/******************************************************************/
// EvalString
eval, err := stl.EvalFull().NewEvaluable(`"hello" + x`)
if err != nil {
    logger.Errorf("EvalFull error: %v", err)
    return
}
stringValue, err := eval.EvalString(context.Background(), map[string]interface{}{"x": " world"})
if err != nil {
    logger.Errorf("Eval error: %v", err)
}
logger.Debugf("Eval = %v", stringValue)
// OutPut: Eval = hello world

/******************************************************************/
// EvalInt
eval, err := stl.EvalFull().NewEvaluable("1 + x")
if err != nil {
    logger.Errorf("EvalFull error: %v", err)
    return
}
intValue, err := eval.EvalInt(context.Background(), map[string]interface{}{"x": 5})
if err != nil {
    logger.Errorf("Eval error: %v", err)
}
logger.Debugf("Eval = %v", intValue)
// OutPut: Eval = 6


```