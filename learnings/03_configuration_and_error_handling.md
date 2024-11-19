# Let's Go - Configuration and error handling

#### Advantages of command line flags vs environment variables
- you get a default value
- you get a `-help` command
- you get type conversion (bool, int, float, ...) (env are always strings)
- you can pass the environment variables into the command line flags
    - ```bash $ export SNIPPETBOX_ADDR=":9999" $ go run ./cmd/web
      -addr=$SNIPPETBOX_ADDR ```

## Structured Logging

### Creating a structured logger
- all structured logger have a **structured logging handler**, that controls
how log entries are formatted and where they are written to
- `logger := slog.New(slog.NewTextHandler(os.Stdout, nil))`
- there is no `log.Fatal()` => `logger.Error(err.Error()) && os.Exit(1)`

### Concurrent logging
- structured loggers are thread-safe and can be savely shared with multiple
goroutines

## Dependency injection
- all handlers are in the same package:
    - create a struct with all the shared dependencies (logger, db pool, ...)
    and register the handlers as methods on the struct
- handlers are in various packages:
    - create another package for the config and use a closure to create the
    handlers:
    - ```Go func ExampleHandler(app *config.Application) http.HandlerFunc {
      return func(w http.ResponseWriter, r *http.Request) { ... ts, err :=
          template.ParseFiles(files...) if err != nil {
          app.Logger.Error(err.Error(), "method", r.Method, "uri",
          r.URL.RequestURI()) http.Error(w, "Internal Server Error",
          http.StatusInternalServerError) return } ... } }
      ```
