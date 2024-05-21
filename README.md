# xltest: A Cross-Language Testing Format

With `xltest`, you write tests in YAML and run them in any language.

## Test format

The `Test` type is described by a [JSON Schema](test-schema.yaml),
written in YAML to allow extensive comments.

Each Test has a short name and a longer description.

A Test specifies an input to some _test function_, and a desired output.
The actual and desired results are compared using a _validation function_,
which is (deep) equality by default.
A Test cannot provide these two functions, because the test is language-agnostic.
Instead, the file containing the Test describes the functions in natural language.

A Test can have subtests, recursively. A Test with subtests need
not have an input or desired output; it can just be container for
other tests.

A test can specify the values of environment variables. The variables
are set for the duration of the test and restored when it finishes.

If the test function signals an error, the test fails by default.
The test can set its `onError` field so that it either succeeds,
in that case, or calls the validation function with the error.

## Example: adding two numbers

Here is the file [`testdata/add.yaml`](testdata/add.yaml):
```
# The function under test returns the sum of two integers.
# Use default validation (equality).

name: add
description: adding two integers
subtests:
  - in: [0, 0]
    want: 0
  - in: [1, 2]
    want: 3
  - in: [-2, 1]
    want: -1
```  

The Go code to run this test looks like this:

```go
func TestAdd(t *testing.T) {
    tst, err := xltest.ReadFile("add.yaml")
    if err != nil {
        t.Fatal(err)
    }
    add := func(args []any) int { return args[0].(int) + args[1].(int) }
    tst.Run(t, add, nil)
}
```

The Javascript code using `node:test` looks like this:

```js
test('add', async (t) = {
  const tst = xltest.readFile('add.yaml'))
  await tst.run(t, (args) => args[0] + args[1])
})
```

The Python code using `unittest` looks like this:

```python
class TestAdd(unittest.TestCase):
    def test_add(self):
        tst = xltest.read_file('add.yaml')
        tst.run(self, lambda args: args[0] + args[1])
``` 

## Implementations

At present, this repo has implementations for Go, Javascript and Python.
The latter two are more proofs of concept than production-ready code, because
I'm not fluent in those languages.

## Contributions

Contributions are welcome, especially implementations for new languages
and improvements to existing ones.

There are deliberately very few features and I want to keep it that way.
Every new feature has to be implemented in every language.


