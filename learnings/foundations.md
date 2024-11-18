# Let's Go - Foundations

## Web application basics
- handlers: Responsible for application logic and writing response header and bodies
- router: Maps URLs to handlers (servemux)
- web server: listens for incoming requests

<!--INFO-->
- `http.ListenAndServe()` always returns an **non-nil** error
    - no check for `if err != nil` needed
- `/` route pattern is a catch-all pattern.

## Routing requests

### Trailing slashes in route patterns
- without trailing `/`: URL must exactly match the pattern, so the corresponding handler is called
- with trailing `/`: URL can have stuff after the `/` and the handler is still called
    - **subtree path pattern** like `/` or `/static/`; it's like the wildcard `/**` or `/static/**`
        + => `/` is a catch-all: match starting `/` and anything or nothing after
- trailing `/` without wildcard behavior add `{$}` => `/{$}` or `/static/{$}`
only matches exactly the pattern

### Additional servemux features
- Request URL paths are automatically sanitized:
    - `/foo/bar/..//baz` => `/foo/baz` with `301 Permanent Redirect`
    - `/foo` => `/foo/` with `301 Permanent Redirect` if `/foo/` is a subtree path pattern

#### Info
- Go initalizes a DefaultServeMux if no servermux is provided to the `ListenAndServe()` function
    - DefaultServeMux is a gloabl variable that can be altered by others / 3rd-party-programs

## Wildcard route patterns
- URL pattern can contain wildcard segments
- use:
    - more flexible routes
    - to pass variables to the application via the URL
- denoted via an wildcard identifier inside `{}` => `/products/{category}/item/{itemID}`
- the content of the wildcards in the request must not be empty
- wildcard must occupy entire segments between the `/` => `/products/c_{category}` is not allowed
- retrieve the value inside the handler using `r.PathValue()`; always returns a string

#### Info
- `http.NotFound(w, r)` replies to the request with an HTTP 404 not found error

### Precedence and conflicts
- 2 or more URL patterns may overlap when using wildcards
- the most specific pattern wins => more specific if it only matches a subset of the other
    - => therefore the order in which the patterns are registered does not matter
- edge case: `/post/new/{id}` and `/post/{author}/latest` both match the pattern `post/new/latest` and both are equally specific
    - => Go's servemux considers those to patterns to *conflict*, and will
    panic at runtime when initalizing the routes
- Remainder wildcards
    - if the route pattern ends with a wildcard and the final wildcard
    identifier ends with `...` then the wildcard will match all remaining
    segements of the request path
        + `/post/{path...}` matches `/post/a`, `/post/a/b`, `/post/a/b/c` like
        a subtree path pattern
        + you can get the value with `r.PathValue("path")`

## Method-based routing
- restrict patterns to only match on specific request methods like `GET`, `POST`, ...
    - `GET /snippet/view/{id}`
    - only one method for a pattern
    - write in UPPERCASE
    - at least one space after
- `GET` method also matches `HEAD` method

- method precedence
    - a pattern with no explicit method matches every method
    - the pattern with the more specific pattern wins

## Customizing responses
- by default every handler sends a HTTP status code `200 OK` plus three
automatic system-generated headers: a Date header, a Content-Length header and
a Content-Type header

### HTTP status codes
- `w.WriteHeader(statuscode int)` to set the status code of the response
- status code can only be sent once
- status code must be set before call to `w.Write()`
- if not status code is set, `w.Write()` automatically calls `w.WriteHeader(200)`

### Customizing headers
- need to change headers before calling `w.WriteHeader()` or `w.Write()`

- `w http.ResponseWriter` is an `io.Writer`
    - `io.WriteString()` and `fmt.Fprint*()` family can be used to write the response

### Content sniffing
- to set the `Content-Type` header, Go sniffs the response body with the
`http.DetectContentType()` function
    - `http.DetectContentType()`: inspects at most the first 512 bytes of data
    and always returns a valid MIME-Type, else it returns
    `application/octet-stream`
- `http.DetectContentType()` cannot detect JSON => returns `text/plain; charset=utf-8`
    - need to set Content-Type manually: `w.Header().Set("Content-Type", "application/json")`

### Manipulating the header map
- type header: `map[string][]string`
- Method signature: `w.Header().Add(key, value string)`
- `Add`: appends a new Header to specified key.
- `Set`: sets the header entries associated with key to the single element value.
- `Get`: gets the first value associated with the given key.
- `Del`: deletes the values associated with key.
- `Values`: returns all values associated with the given key.

### Header canonialization
- headers are canonialized automatically
- to avoid this alter the header map directly =>
` w.Header()["X-XSS-Protection"] = []string{"1; mode=block"}`

## HTML templating and inheritance
- `ts, err := template.ParseFiles(filenames... string)`: parses multiple HTML
templates and combines them into a single template object.
- `ts.Execute(w, nil)`: generates the output of the template and writes it to
w. Optionally, data can be passed into the template for dynamic content
rendering.

#### Info
- `http.Error(io.Writer, msg, status code)`: lightweight helper functions that
response with a plain text error message and status code
    - it does not end the request; need to make sure no furher writes are done
    to w => `return` after
- pattern:
    ```Go
    if err != nil {
        log.Print(err.Error())
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    ```

### Template composition
- `{{define "base"}}...{{end}}`: to define a distinct named template
- `{{template "title" .}}`: to invoke other named templates inside a template
(called title) at a particular location in the html
    - `.`: represents any dynamic data that is passed to the invoked template

- only one template:
    - `template.ParseFiles()` => `ts.Execute()`
- with multiple templates and therefore one base template:
    - `template.ParseFiles()` => `ts.ExecuteTemplate(w, "base", nil)`

##### Info
- block action: `{{block}}...{{end}}` like `{{template}}` action, but with
optional default value if invoked template does not exist

## Serving static files
- to add static css, images files and javascript

### The http.FileServer handler
- `net/http` package ships with a built-in `http.FileServer` handler which you
can use to serve files over HTTP from a specific directory.
- `http.FileServer(root FileSystem) Handler`
    - `FileSystem` (Interface): `Open(name string) (File, error)`
        + to use the operating system's file systems implementation: `http.Dir(path)`
        + `http.Dir type string`
    - `Handler` (interface): `ServeHTTP(ResponseWriter, *Request)`
- example
    ```Go
    fileServer := http.FileServer(http.Dir("./ui/static/"))
    mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
    ```
    - pattern is `/static/` with trailing slash to be a subtree path pattern like a wildcard
- Why use `StripPrefix("/static")`?
    - the filesystem is rooted under `./ui/static/`; `./ui/static/` is the
    root of the filesystem
    - if `/static` was not striped, the filesystem would search under
    `./ui/static/static` for the files and does not find them there

### File server features and functions
- sanitizes all request paths by running them through the `path.Clean()` function
    - stops directory traversal attacks
- `Range requests` are fully supported.
    - support for resumable downloads
    - `curl -i -H "Range: bytes=100-199" --output - localhost:4000/static/img/logo.png`
- `Last-Modified` and `If-Modified-Since` headers are supported.
    - if file has not changed since last user request, user gets an `304 Not Modified` status code
- `Content-Type` is automatically set from the file extension using `mime.TypeByExtension()`
    - `mime.AddExtesionType()`: to add custom extensions

### Serving single files
- `http.ServeFile(w, r, path)`
- `http.ServeFile()` does not sanitize the file path; If the file path is
user-provided, ensure you sanitize it using `filepath.Clean()` before passing it
to the function.

### Disabling directory listings
- easiest way is to add a blank `index.html` file to that directory
- then the blank `index.html` file is server and the user gets a `200 OK`
- `find ./ui/static -type d -exec touch {}/index.html \;`
- a better solution is to create a custom `http.FileServer` implementation
returning a `os.ErrNotExist` error ([see](https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings))

## The http.Handler interface
- interface with `ServeHTTP(ResponseWriter, *Request)`
    - it other terms it must be an object with the exact `ServeHTTP` method
- example:
    ```Go
    type home struct {}

    func (h *home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("This is my home page"))
    }

    ```
    - is the easiest Handler
- usage: `mux.Handle("/", &home{})`
- shorter: `mux.Handle("/", http.HandlerFunc(home))`
    - adapter pattern: adds a `ServeHTTP` method the the `home` method; the
    `ServeHTTP()` just calls `home()`
- even shorter: `mux.HandleFunc("/", home)`
    - transforms a function to a handler and registers it in one step

### Chaining handlers
- `http.ListenAndServe(addr string, handler http.Handler)`, but we pass in a `servemux`
    - `http.ServeMux` also has a `ServeHTTP` method
- `servemux` is a special kind of handler's `ServeHTTP()` that receives the
request and calls the corresponding handler based on the URL

### Requests are handled concurrently
- **all incoming HTTP requests are served in their own goroutine.**
- be aware of (and protect against) race conditions when accessing shared
ressources from your handlers

#### Curl
- `-i, --include`: Include the HTTP response headers in the output
- `-d, --data`: To send data. Automatically uses POST as method
- `-X, --request`: Change the method to use when starting the transfer
- `-L, --location`: automatically follows redirects
