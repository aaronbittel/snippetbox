# Processing forms

## Parsing form data
1. use `r.ParseForm()` to parse the request body.
    - this populates the request's `r.PostForm` map
2. use `r.PostForm.Get()` to retrieve form values by key
    - e.g. `r.PostForm.Get("title")`

### The PostFormValue method
- is a shortcut that does `r.ParseForm()` and `r.PostForm.Get()` in one step,
but it silently ignores all errors `r.ParseForm()` returns
    => DO NOT USE IT

### Multiple-value fields
- `r.PostForm.Get()` only retrieves the first value of a form field
- if a form field returns multiple values use the underlying map
(`map[string][]string`) directly to retrieve the form values

### Limiting form size
- default limit with POST-method: 10MB
- exception: form has `enctype="multipart/form-data"` attribute => no limit
- change the default limit using `http.MaxBytesReader()`:
    ```Go
        r.Body = http.MaxBytesReader(4096)
        err := r.ParseForm()
    ```
    - first 4096 bytes wil be read
- if the limit is reached the underlying TCP connection is closed

### Query string parameters
- using GET as form method includes the data in the URL query string parameters
    - e.g. `/foo/bar?title=value&content=value`
- retrieve the data using `r.URL.Query().Get()`

## Validating form data
- [validation snippets](https://www.alexedwards.net/blog/validation-snippets-for-go)
