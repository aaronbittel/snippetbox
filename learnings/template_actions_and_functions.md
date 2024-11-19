# Templating Actions
- `{{define "base"}}...{{end}}`: to define a **distinct named template
called "base"**
- `{{template "main" .}}`: to invoke another template (called "main") at a
particular location in the HTML
    - the "." at the end represent any dynamic data we want to pass into
    the invoked template
- `{{block "sidebar" .}}...{{end}}`: acts like `{{template}}` action, execpt it
allows to specify default content
- `{{if .Foo}} C1 {{else}} C2 {{end}}`: If `.Foo` is not empty than render C1
otherwise C2
- `{{with .Foo}} C1 {{else}} C2 {{end}}`: If `.Foo` is not empty than set . to
the value of `.Foo` and render C1 else C2
    - if "." is the templateData struct (`type struct {Snippet:
    models.Snippet}`), then I can change the value of "." for that segment:
    `{{with .Snippet}}...{{end}}`, so for that segment the "." refers to
    `Snippet` instead of `templateData`
- `{{range .Foo}} C1 {{else}} C2 {{end}}`: If `.Foo.len()` > 0 render C1 for
each instance else C2
    - `.Foo` must be Array, Slice, Map or Channel

- `{{else}}` is optional
- `{{with}}` and `{{range}}` both change the value of "."!!

# Templating Functions
- `{{eq .Foo .Bar}}`: Yields true if .Foo is equal to .Bar
- `{{ne .Foo .Bar}}`: Yields true if .Foo is not equal to .Bar
- `{{not .Foo}}`: Yields the boolean negation of .Foo
- `{{or .Foo .Bar}}`: Yields .Foo if .Foo is not empty; otherwise yields .Bar
- `{{index .Foo i}}`: Yields the value of .Foo at index i. The underlying type
of .Foo must be a map, slice or array, and i must be an integer value.
- `{{printf "%s-%s" .Foo .Bar}}`: Yields a formatted string containing the .Foo
and .Bar values. Works in the same way as fmt.Sprintf().
- `{{len .Foo}}`: Yields the length of .Foo as an integer.
- `{{$bar := len .Foo}}`: Assign the length of .Foo to the template variable
$bar
