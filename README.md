Package `inflect` can pluralize ("cat" => "cats", "man" => "men")
or singularize ("cats" => "cat", "men" => "man") English language nouns.

API docs are [here](https://godoc.org/github.com/kjk/inflect).

Usage:
```go
import "github.com/kjk/inflect"

inflect.ToPlural("man") // "men"
inflect.ToPlural("men") // "men"

inflect.ToSingular("men") // "man"
inflect.ToSingular("man") // "man"

inflect.IsPlural("cats") // true
inflect.IsPlural("cat")  // false

inflect.IsSingular("cat")  // true
inflect.IsSingular("cats") // false
```

This is a Go port of https://github.com/blakeembrey/pluralize
