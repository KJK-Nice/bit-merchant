# Port Rules (PORT-01..06)

## PORT-01: Handler Struct Holds Application (WARNING)

HTTP and gRPC handler structs MUST hold `app.Application` and delegate to it. They are thin wrappers.

```go
// ports/http.go
type HttpServer struct {
    app app.Application
}

// ports/grpc.go
type GrpcServer struct {
    app app.Application
}
```

---

## PORT-02: Error Mapping (WARNING)

Ports MUST map application errors to protocol-specific responses. They must NOT leak internal error details.

**HTTP — using httperr helper:**
```go
func (h HttpServer) MakeHourAvailable(w http.ResponseWriter, r *http.Request) {
    err = h.app.Commands.MakeHoursAvailable.Handle(r.Context(), command.MakeHoursAvailable{...})
    if err != nil {
        httperr.RespondWithSlugError(err, w, r)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

**The httperr mapper:**
```go
func RespondWithSlugError(err error, w http.ResponseWriter, r *http.Request) {
    slugError, ok := err.(errors.SlugError)
    if !ok {
        InternalError("internal-server-error", err, w, r)
        return
    }
    switch slugError.ErrorType() {
    case errors.ErrorTypeAuthorization:
        Unauthorised(slugError.Slug(), slugError, w, r)  // 401
    case errors.ErrorTypeIncorrectInput:
        BadRequest(slugError.Slug(), slugError, w, r)     // 400
    default:
        InternalError(slugError.Slug(), slugError, w, r)  // 500
    }
}
```

**gRPC — using status codes:**
```go
func (g GrpcServer) ScheduleTraining(ctx context.Context, req *trainer.UpdateHourRequest) (*empty.Empty, error) {
    if err := g.app.Commands.ScheduleTraining.Handle(ctx, command.ScheduleTraining{...}); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &empty.Empty{}, nil
}
```

---

## PORT-03: Auth Extracted from Context (WARNING)

Authentication/authorization data MUST be extracted from the request context using a shared auth package, NOT parsed directly in the handler.

**Correct:**
```go
func (h HttpServer) MakeHourAvailable(w http.ResponseWriter, r *http.Request) {
    user, err := auth.UserFromCtx(r.Context())
    if err != nil {
        httperr.RespondWithSlugError(err, w, r)
        return
    }
    if user.Role != "trainer" {
        httperr.Unauthorised("invalid-role", nil, w, r)
        return
    }
    // ... delegate to app
}
```

**Wrong:**
```go
func (h HttpServer) MakeHourAvailable(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")          // VIOLATION: parsing auth in handler
    claims, err := jwt.Parse(token, keyFunc)         // VIOLATION: JWT logic in port
    // ...
}
```

---

## PORT-04: No Business Logic in Ports (CRITICAL)

Port handlers MUST only:
1. Parse/decode the request
2. Extract auth from context
3. Construct command/query struct
4. Call `app.Commands.X.Handle()` or `app.Queries.X.Handle()`
5. Map the result/error to a response

They MUST NOT contain:
- Domain validation logic
- Business rule checks
- Direct database calls
- State manipulation

**Check:** Port files should only import `app/`, `app/command/`, `app/query/`, and infrastructure packages (HTTP, gRPC, auth). They should NOT import `domain/` directly (except for response mapping types).

---

## PORT-05: Response Model Mapping (INFO)

Response transformation SHOULD be in separate mapping functions, not inline in handlers.

```go
// Mapping function
func dateModelsToResponse(models []query.Date) []Date {
    var dates []Date
    for _, m := range models {
        dates = append(dates, Date{
            Date:  m.Date,
            Hours: hourModelsToResponse(m.Hours),
        })
    }
    return dates
}

// Handler uses it cleanly
func (h HttpServer) GetTrainerAvailableHours(w http.ResponseWriter, r *http.Request, params GetTrainerAvailableHoursParams) {
    dateModels, err := h.app.Queries.TrainerAvailableHours.Handle(r.Context(), query.AvailableHours{
        From: params.DateFrom,
        To:   params.DateTo,
    })
    if err != nil {
        httperr.RespondWithSlugError(err, w, r)
        return
    }
    dates := dateModelsToResponse(dateModels)
    render.Respond(w, r, dates)
}
```

---

## PORT-06: No Unimplemented Embedding in gRPC Servers (CRITICAL)

gRPC server structs MUST NOT embed `Unimplemented*Server` structs. Omitting the embed enforces **compile-time interface compliance** — if a new RPC is added to the proto definition, the code will fail to compile until the method is explicitly implemented.

Embedding `Unimplemented*Server` silently returns "unimplemented" at runtime for missing methods, hiding broken contracts until a request hits the missing endpoint in production.

**Correct:**
```go
type GrpcServer struct {
    app app.Application
}
// Compile error if any RPC method from TrainerServiceServer is missing.
```

**Wrong:**
```go
type GrpcServer struct {
    trainer.UnimplementedTrainerServiceServer  // VIOLATION: hides missing methods at compile time
    app app.Application
}
```

**Check:** Scan all structs in `ports/grpc.go` for embedded `Unimplemented*Server` fields. Any match is a CRITICAL violation.

**Proto generation:** When generating gRPC code, use `require_unimplemented_servers=false` to keep the interface strict:
```
protoc --go-grpc_out=require_unimplemented_servers=false:. *.proto
```
