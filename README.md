`inflect` package is a Go port of https://github.com/blakeembrey/pluralize

It can inflect english nouns i.e. pluralize ("cat" => "cats", "man" => "men")
or singularize ("cats" => "cat", "men" => "man").

Usage:
```go
import "github.com/kjk/inflect"

inflect.ToPlural("man") // "men"

inflect.ToSingular("men") // "man"

inflect.IsPlural("cats") // true

inflect.IsSingular("cat") // true
```
