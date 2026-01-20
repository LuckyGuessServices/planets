
# TODO

- [MVP](#mvp)
- [Future development](#future-development)
- [Unscheduled optional tasks](#unscheduled-optional-tasks)

## MVP

1. - [ ] Ensure the database contents remain the same at the beginning of each test. Options:
    * (preferable) Rollback. See https://github.com/DATA-DOG/go-txdb

      Or implement an own wrapper for transactions management.
    * (backup plan) `tmpfs` plus `CREATE DATABASE "X" WITH TEMPLATE "Y"`
1. - [ ] Catch all handlers' panics and turn into valid responses.
1. - [ ] Add `mercuryretrogradeapi.com` API client. See [docs](https://mercuryretrogradeapi.com/about.html).
    1. - [ ] Implement a method for the only endpoint `/` and its optional parameter `date`
         (but consider it as a mandatory parameter).
    1. - [ ] Cover with _mock_ autotests:
        * 200, expected response
        * 200, unexpected response (alert about even safe changes like a single extra field)
        * 4xx/5xx responses
        * no response / time-out
    1. - [ ] (optionally) Add _manually launched_ real autotests (maybe as a separate "application").
1. - [ ] Pick and test an ORM library. Candidate: https://github.com/ent/ent
1. - [ ] Add `on-date` GET endpoint. Utilize the API client created on the previous step.
    1. - [ ] Definition:
        * Request parameters:
            * date: `YYYYMMDD` or any automatically recognizable format
        * Response:
            * array of objects, each: (int) planet_id => (bool) is_retrograde
    1. - [ ] If there is no data locally, request the external API, store the data locally and then return the data.
        * Add a `TODO` comment to store data in background in the future.
    1. - [ ] Consider edge cases:
        * Invalid date.
        * No data in the local storage, external APIs are unavailable (no response or time out).
    1. - [ ] Cover `on-date` endpoint with autotests.
        1. - [ ] Mock external API(s).
        1. - [ ] Try to explicitly forbid external connections - if you add a new external API and forget to update
             autotests, the latter will fail while trying to request an external IP.
        1. - [ ] Cover cases: data from external API, data from the local storage, invalid date, no data available.

## Future development

1. - [ ] Add versions to API like `/api/vX.Y`
1. - [ ] Simplify writing API tests:
     allow to specify just then ending part of an endpoint URI (like `/` instead of `/api/`),
     but ensure tests load the whole router (not just the API sub-router).
1. - [ ] Document the service's REST API with OpenAPI.
1. - [ ] Add a queue manager / message broker / etc. to request external API(s) in background.

    Until data is received, API server should return something like "please, try again later" responses
    (in a way that is easy enough to automate retries).
    
## Unscheduled optional tasks

1. - [ ] Add gRPC (https://grpc.io/) as a second API (with the same functionality; just for education)
     Or implement it as a single API server in another microservice.
1. - [ ] Upgrade [luglog.go](../src/internal/luglog/luglog.go) calls:
    1. - [ ] Implement logging levels.
    1. - [ ] Add an option to additionally or exclusively log to files.
1. - [ ] Add own Logger to `goose`: implement `goose.Logger` and apply it to `goose.SetLogger()`.
1. - [ ] Cover [api.go](../src/internal/api/api.go) general functions with tests.
    1. - [ ] Stripping URI prefixes from API endpoints' URIs.
    1. - [ ] Edge cases with errors.
