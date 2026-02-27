
# TODO

- [Baseline](#baseline)
- [Unscheduled optional tasks](#unscheduled-optional-tasks)

## Baseline

1. - [ ] Add a lint check to ensure every test package contains a standardized TestMain call.
1. - [ ] Add a queue manager / message broker / etc. to request external API(s) in background.
     - [ ] Until data is received, API server should return something like "please, try again later" responses
     (in a way that is easy enough to automate retries).
1. - [ ] Add caching for API requests (full or granular).
1. - [ ] Add versions to API like `/api/vX.Y`
1. - [ ] Document the service's REST API with OpenAPI.
1. - [ ] Document API error codes.

## Unscheduled optional tasks

1. - [ ] Alter mocked time for all tests:
     "a second later" sleep should produce cases of a leap second and / or an extra hour (winter/summer).
1. - [ ] Try replacing mercury API with own retrograde state calculation.
1. - [ ] Try to forbid external connections automatically (until allowed explicitly): if you add a new external
     API and forget to update autotests, the latter will fail while trying to request an external IP.
1. - [ ] Add another external source: https://freeastrologyapi.com/api-reference/planets
    * Advantage: retrograde state (and some other parameters) for all Sol planets.
    * Caveat: a free token limits to 50 requests per day.
1. - [ ] Add gRPC (https://grpc.io/) API as a replacement for REST API
     (with the same functionality; just for education).
1. - [ ] Upgrade [luglog.go](../src/internal/luglog/luglog.go) calls:
    1. - [ ] Implement logging levels.
    1. - [ ] Log milli- or nanoseconds.
    1. - [ ] Add an option to additionally or exclusively log to files.
1. - [ ] Add own Logger to `mercury`: log each request (URL + query), response (status code, limited body).
    * Optionally, prepare a more universal solution (for any API client) located in `integration` package.
1. - [ ] Add own Logger to `goose`: implement `goose.Logger` and apply it to `goose.SetLogger()`.
1. - [ ] (with a high-load generator service) Add metrics. 
1. - [ ] Make it possible for `databases.Core()` to return `pgxpool.Pool` instead of `sql.DB`.
     But it should work with `DATA-DOG/go-txdb` as well (e.g. via a common interface).

     The caveat: it should work with `DATA-DOG/go-txdb` as well. To achieve this, try implementing a common interface
     (as a return type hint). Otherwise, replace `DATA-DOG/go-txdb` with your own implementation based on `pgxpool`
     (see [this post](https://github.com/jackc/pgx/issues/697#issuecomment-604035545)).
1. - [ ] Cover [API](../src/internal/api) general functions (and edge cases with errors) with tests.
1. - [ ] `mercury` API: add _manually launched_ real autotests (maybe as a separate "application").
