# Dynamic HTML templates

## Displaying dynamic data
- any exported fields of the passed data to `ts.ExecuteTemplate()` can be
accessed using the "." notation e.g. {{.Title}
    - only one object can be passed on the `ts.ExecuteTemplate()` method
    - => create a struct holding all the data (structs) and pass in that
    "template struct"

### Dynamic content escaping
- `html/template` package automatically escapes any data that is yielded
between `{{ }}` tags, preventing cross-site scripting (XSS) attacks

### Calling methods
- you can call exported methods on the data (if they only return `value` or
`(value, error)`)
- e.g. `<span>{{.Snippet.Created.AddDate 0 6 0}}</span>`
    - no use of `()` and parameters use " " instead of ","

## Template actions and functions
- See `template_actions_and_functions.md` for a list

### Combining functions
- use `()` to surround functions and their arguments
    - e.g. `{{if (gt (len .Foo) 99)}} C1 {{end}}`
    - e.g. `{{if (and (eq .Foo 1) (le .Bar 20))}} C1 {{end}}`

### Controlling loop behavior
- use `{{continue}}` or {{break}}
- e.g.
    ```html
    {{range .Foo}}
        // Skip this iteration if the .ID value equals 99.
        {{if eq .ID 99}}
            {{continue}}
        {{end}}
        // ...
    {{end}}
    ```
- e.g.
    ```html
    {{range .Foo}}
        // End the loop if the .ID value equals 99.
        {{if eq .ID 99}}
            {{break}}
        {{end}}
        // ...
    {{end}}
    ```

## Catching runtime errors
- adding dynamic behavior to HTML templates => risk for runtime errors
    - problem: user gets a 200 OK, but only a half-completed HTML page
    althrough the program has thrown an error
    - solution: make the HTML rendering a two-step-process.
        + First: Write to a buffer if that succeeds
        + Second: Write from buffer to http.ResponseWriter

## Common dynamic data
- add in to the templateData struct
- create a `newTemplateData()` that returns all the common base data
- use it in e.g. the "base" template

## Custom template functions
1. create `template.FuncMap` object containing the custom function
    - `type FuncMap map[string]any`: defines the mappings from names to
    functions
    - each function must have either one return value or two (with the second
    being error)
2. use `template.Funcs()` to register the map to the template's function map.
   This must be done  before parsing the template

### Pipelining
- instead of `<time>Created: {{humanDate .Created}}</time>`
- you can use pipelining: `<time>Created: {{.Created | humanDate}}</time>`
- `<time>{{.Created | humanDate | printf "Created: %s"}}</time>`
