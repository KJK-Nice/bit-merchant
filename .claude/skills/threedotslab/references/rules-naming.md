# Naming Rules

## Strict Naming Convention Table

| Pattern | Convention | Example |
|---------|-----------|---------|
| Entity constructor | `New{Type}(args...) (*Type, error)` | `NewTraining(...)`, `NewAvailableHour(...)` |
| Panic constructor | `MustNew{Type}(args...) Type` | `MustNewFactory(...)`, `MustNewUser(...)` |
| DB reconstruction | `Unmarshal{Type}FromDatabase(...)` | `UnmarshalHourFromDatabase(...)` |
| Value from string | `New{Type}FromString(s string) (Type, error)` | `NewAvailabilityFromString(...)` |
| Command struct | Imperative verb + noun (PascalCase) | `ScheduleTraining`, `CancelTraining`, `MakeHoursAvailable` |
| Query struct | Noun phrase (PascalCase) | `AvailableHours`, `HourAvailability`, `AllTrainings` |
| Handler type (exported) | `{ActionName}Handler` | `ScheduleTrainingHandler`, `CancelTrainingHandler` |
| Handler struct (unexported) | `{actionName}Handler` | `scheduleTrainingHandler`, `cancelTrainingHandler` |
| Handler constructor | `New{ActionName}Handler(...)` | `NewScheduleTrainingHandler(...)` |
| Adapter type | Technology suffix | `FirestoreHourRepository`, `MySQLHourRepository`, `MemoryHourRepository` |
| Adapter constructor | `New{Tech}{Entity}Repository(...)` | `NewFirestoreHourRepository(...)` |
| DB model (SQL) | Tech prefix, unexported | `mysqlHour`, `postgresTraining` |
| DB model (NoSQL) | `{Entity}Model` (exported for tags) | `TrainingModel`, `DateModel` |
| Sentinel errors | `Err{Name}` | `ErrNotFullHour`, `ErrHourNotAvailable` |
| Typed errors | `{Condition}Error` | `TooDistantDateError`, `NotFoundError` |
| Zero check | `IsZero() bool` | `Availability.IsZero()`, `Factory.IsZero()` |
| Application struct | `Application` in `app/` package | `app.Application` |
| App sub-structs | `Commands`, `Queries` | `app.Commands`, `app.Queries` |
| Composition root | `NewApplication(...)` in `service/` | `service.NewApplication(ctx)` |
| gRPC client adapter | `{Service}Grpc` | `TrainerGrpc`, `UsersGrpc` |
| Read model interface | `{Query}ReadModel` | `AvailableHoursReadModel` |

## CRUD-to-Domain-Language Mapping

CRUD terms are **forbidden** in domain code, commands, queries, and API endpoints. Use domain-specific language instead.

| CRUD Term | Replacement Options | Example |
|-----------|-------------------|---------|
| Create | Schedule, Register, Place, Submit, Open, Enroll | `ScheduleTraining`, not `CreateTraining` |
| Read | *(use noun phrase queries)* | `AvailableHours`, not `GetHours` |
| Update | Approve, Reject, Reschedule, Move, Modify, Assign | `ApproveReschedule`, not `UpdateTraining` |
| Delete | Cancel, Archive, Revoke, Close, Withdraw | `CancelTraining`, not `DeleteTraining` |
| Get | *(avoid as prefix)* | `HourAvailability`, not `GetHourAvailability` |
| Set | *(use specific verb)* | `MakeAvailable`, not `SetAvailability` |
| List | *(use noun phrase)* | `AllTrainings`, not `ListTrainings` |
| Fetch | *(avoid entirely)* | Use noun phrase queries |

## Check Procedure

1. Scan all type declarations and function names
2. Flag any use of Create/Read/Update/Delete/Get/Set/List/Fetch in:
   - Command struct names
   - Query struct names
   - Domain entity method names
   - Handler type names
3. Severity: CRITICAL for command/query names, WARNING for methods
