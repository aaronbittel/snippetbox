# Middleware
- shared functionality to use for many (even all) requests
- e.g. log incoming requests, compress responses, check cache before handing it
over to handlers

## How middleware works
- insert another handler into the handler chain
- execute some code
- call the next handler's `ServeHTTP()` method

### Pattern
- ```Go
    func myMiddleware(next http.Handler) http.Handler {
        fn := func(w http.ResponseWriter, r *http.Request) {
            // TODO: Execute our middleware logic here...
            next.ServeHTTP(w, r)
        }

        return http.HandlerFunc(fn)
    }
    ```
    - the `fn` function does what the middleware wants to do
    - at the end the `next` http.Handler's `ServeHTTP` method is called to
    return control
    - `myMiddleware` function returns a `http.Handler` by using
    `http.HandlerFunc(fn)` adapter to turn a function into a `http.Handler`

### Simplifiying the middleware
- ```Go
    func myMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // TODO: Execute our middleware logic here...
            next.ServeHTTP(w, r)
        })
    }
    ```

### Positioning the middleware
- before the servemux => middleware gets executed on every request; e.g. for
request logging
- after the servemux => middleware gets exectuted only for specific routes;
e.g. for authorization

## Setting common headers
- needs to be called before servemux

### Flow of control
- when the last handler in the chain returns, the control is passed back in
reverse order
    - commonHeaders → servemux → application handler → servemux → commonHeaders
- everything **before** the call of `next.ServeHTTP` is done the "way down" the
chain
- everything **after** the call of `next.ServeHTTP` is done the "way up" the
chain

### Early return
- calling `return` before `next.ServeHTTP` will cause the chain to stop and
flow back up
    - e.g. a request that did not pass the authentication middleware

## Panic recovery
- every request is handled in a seperate goroutine
- if a goroutine panics, the connection is closed, but the server keeps running
- buf if a `panic()` occurs in one of the handlers the user receives a `Empty
reply from server` response
=> install a `recoverPanic()` middleware
    - all deferred functions are called when the stack is being unwound
    following the panic
    - ```Go
        func (app *application) recoverPanic(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                defer func() {
                    if err := recover(); err != nil {
                        w.Header().Set("Connection", "close")
                        app.serverError(w, r, fmt.Errorf("%s", err))
                    }
                }()

                next.ServeHTTP(w, r)
            })
        }
        ```
        + Setting `Connection: close` header on the response triggers Go's HTTP
        server to automatically close the connection after sending the response
        + if HTTP/2 is used Go will automatically remove the `Connection:
        close` header and send a `GOAWAY` frame

### Panic recovery in background goroutines
- if a handler spins up a goroutine (and this goroutine panics), the
`recoverPanic` middleware does not recover that panic
    - `recover()` will only recover panics in the same goroutine
- use:
    ```Go
    func (app *application) myHandler(w http.ResponseWriter, r *http.Request) {
        ...

        // Spin up a new goroutine to do some background processing.
        go func() {
            defer func() {
                if err := recover(); err != nil {
                    app.logger.Error(fmt.Sprint(err))
                }
            }()

            doSomeBackgroundProcessing()
        }()

        w.Write([]byte("OK"))
    }

    ```

## Composable middleware chains
