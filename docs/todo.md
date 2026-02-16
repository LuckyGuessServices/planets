
# TODO

- [MVP](#mvp)
- [Future development](#future-development)
- [Unscheduled optional tasks](#unscheduled-optional-tasks)

## MVP

1. - [ ] Add `mercuryretrogradeapi.com` API client. See [docs](https://mercuryretrogradeapi.com/about.html).
    1. - [ ] Implement a method for the only endpoint `/` and its optional parameter `date`
         (but consider it as a mandatory parameter).
    1. - [ ] Cover with _mock_ autotests:
        * 200, expected response
        * 200, unexpected response (alert about even safe changes like a single extra field)
        * 4xx/5xx responses
        * no response / time-out
    1. - [ ] (optionally) Add _manually launched_ real autotests (maybe as a separate "application").
1. - [ ] Add `on-date` GET endpoint. Utilize the local database model and the API client created on the previous steps.
    1. - [ ] Definition:
        * Request parameters:
            * date: `YYYYMMDD` or any automatically recognizable format
        * Response:
            * array of objects, each: (int) planet_index => (bool) is_retrograde
    1. - [ ] Try retrieving data from the local database. If there is no data locally, request the external API,
         store the data locally and then return the data.
    1. - [ ] Consider edge cases:
        * Invalid date.
        * No data.
        * External APIs are unavailable (no response or time out).
    1. - [ ] Cover `on-date` endpoint with autotests.
        1. - [ ] Mock external API(s).
        1. - [ ] Try to explicitly forbid external connections - if you add a new external API and forget to update
             autotests, the latter will fail while trying to request an external IP.
        1. - [ ] Cover cases:
            * Typical success.
            * Invalid parameters.
            * Data gathered from local storage (if stored) or external API (plus storing locally) or no data available.

## Future development

1. - [ ] Add versions to API like `/api/vX.Y`
1. - [ ] Document the service's REST API with OpenAPI.
1. - [ ] Document API error codes.
1. - [ ] Add a queue manager / message broker / etc. to request external API(s) in background.

    Until data is received, API server should return something like "please, try again later" responses
    (in a way that is easy enough to automate retries).
1. - [ ] Add caching for API requests (full or granular).

## Unscheduled optional tasks

1. - [ ] Add a lint check to ensure every test package contains a standardized TestMain call.
1. - [ ] Alter mocked time for all tests:
     "a second later" sleep should produce cases of a leap second and / or an extra hour (winter/summer).
1. - [ ] Add gRPC (https://grpc.io/) as a second API (with the same functionality; just for education)
     Or implement it as a single API server in another microservice.
1. - [ ] Upgrade [luglog.go](../src/internal/luglog/luglog.go) calls:
    1. - [ ] Implement logging levels.
    1. - [ ] Log milli- or nanoseconds.
    1. - [ ] Add an option to additionally or exclusively log to files.
1. - [ ] Add own Logger to `goose`: implement `goose.Logger` and apply it to `goose.SetLogger()`.
1. - [ ] Make it possible for `databases.Core()` to return `pgxpool.Pool` instead of `sql.DB`.
     But it should work with `DATA-DOG/go-txdb` as well (e.g. via a common interface).

     The caveat: it should work with `DATA-DOG/go-txdb` as well. To achieve this, try implementing a common interface
     (as a return type hint). Otherwise, replace `DATA-DOG/go-txdb` with your own implementation based on `pgxpool`
     (see [this post](https://github.com/jackc/pgx/issues/697#issuecomment-604035545)).
1. - [ ] Cover [API](../src/internal/api) general functions (and edge cases with errors) with tests.
