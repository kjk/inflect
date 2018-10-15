`inflect` package is a Go port of https://github.com/blakeembrey/pluralize

It can inflect english nouns i.e. pluralize ("cat" => "cats", "man" => "men")
or singularize ("cats" => "cat", "men" => "man").

Usage:
```go
import "github.com/kjk/inflect"

s := inflect.ToPlural("man")
// s == "men"

s = inflect.ToSingular("men")
// s == "man"

inflect.IsPlural("cats") // true
inflect.IsSingular("cat") // true
```
