# Server and security improvements

## The http.Server struct
- `http.ListenAndServe()` is a shortcut for starting a server
-   ```Go
        srv := http.Server {
            Addr: ":4000",
            Handler: app.routes(),
        }

        err := srv.ListenAndServe()
        ```
    - need to configure `Addr` and `Handler` because
    `http.Server.ListenAndServe()` does not take them as parameters
    - has more possibilites for configuration

## The server error log
- by default `http.Server` logs to the standard logger
- when using the `slog.Logger` need to redirect the log messages
    - ```Go
        http.Server {
            ErrorLog: slog.NewLogLogger(slog.Handler(), slog.LevelError)
        }
        ```
    - this writes all the log messages as error level to the `slog.Logger`

## Generating a self-signed TLS certificate
- to make the server use HTTPS rather than HTTP and handle the connection
across TLS (Transport Layer Security)
- before we can use HTTPS, we need to generate a TLS certificate
    - for production servers use [Let's Encrypt](https://letsencrypt.org/)
- `go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost`
creates a `cert.pem` and `key.pem` files
    1. generates a 2048 RSA key pair (cryptographically secure private and
       public key)
    2. stores the private key in `key.pem` and generates a self-signed TLS
       certificate for the host `localhost` containing the public key which is
       stores in `cert.pem`
       + add the `tls/` directory to `.gitignore`
- browser will first warn about self-signed TSL certificate

## Running a HTTPS server
- use `srv.ListenAndServeTLS()` instead of `stv.ListenAndeServe()` and provide
the path to the `cert.pem` and `key.pem` files
- also set `sessionManager.Cookie.Secure = true`

### HTTP requests
- all HTTP requests, the server will send a `400 Bad Request` with the message
`Client sent HTTP request to an HTTPS server`. Nothing will be logged

### HTTP/2 connections
- Go will automatically upgrade the connection to use HTTP/2 if the client
supports it
    - rundown of HTTP/2: [GoSF meetup talk](https://www.youtube.com/watch?v=FARQMJndUn0)

## Configuring HTTPS settings
- use `tls.Config` struct and provide it to the `http.Server`
- restricting elliptic curves in the TLS handshake to `tls.CurveP256` and
`tls.X25519` for better performance (both have assembly implementations)
- ```Go tlsConfig := &tls.Config{
        CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
    }```

## Connection timeouts
- `IdleTimeout`
    - by default Go enables keep-alive on all accepted connections
        + client can reuse the same connection for multiple requests without
        repeating the TLS handshake
    - by default keep-alive connections are automatically closed after a couple
    of minutes (depending on OS), helps to clear-up connections where the user
    has disappeated unexpectetly (power cut)
    - cannot increase this default but decrease it using `IdleTimeout`
    - automatically close a connection after this time of inactivity
- `ReadTimeout`
    - if the response body is still read after this timeout is passed the
    connection is automatically closed
        + it a hard closure => client receives no response
    - mitigate the risk of slow-client-attacks
    - if you set `ReadTimeout`, but no `IdleTimeout`, `IdleTimeout` is set to
    the specified `ReadTimeout`
- `WriteTimeout`
    - connection will be closed if the server tries to write to it after the
    specified time has elapsed since the request was accepted.
        - set the `WriteTimeout` to be larger than the `ReadTimeout`
        - both are measured as **time after the request was accepted**, and the
        body must be read before the response is written

### The MaxHeaderBytes setting
- `http.Server` includes a `MaxHeaderBytes` field to control the maximum number
of bytes the server reads when parsing the request headers
- default of 1MB
if `MaxHeaderBytes` is exceeded the user will get a `431 Request Header Fields
Too Large`
- Go always adds an **additional 4096 bytes** of headroom to the figure
