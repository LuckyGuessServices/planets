
# TODO

- [MVP](#mvp)
- [Misc](#misc)

## MVP

1. - [ ] Add database migrations tool. Minimal requirements:
    * Migration version control mechanism.
    * Go-based migration files.
1. - [ ] Ensure the database contents remain the same at the beginning of each test. Options:
    * Rollback. See https://github.com/DATA-DOG/go-txdb
    * `tmpfs` plus DB template duplication.
1. - [ ] Catch all handlers' panics and turn into valid responses.
1. - [ ] Base the service's API on OpenAPI.
1. - [ ] Add `mercuryretrogradeapi.com` API client. See [docs](https://mercuryretrogradeapi.com/about.html).
    1. - [ ] Implement a method for the only endpoint `/` and its optional parameter `date`
         (but consider it as a mandatory parameter).
    1. - [ ] Cover it with _manually launched_ autotests - should not be launched for each commit.
1. - [ ] Add `on-date` GET endpoint. Utilize the API client created on the previous step.
    1. - [ ] Definition:
        * Request parameters:
            * date: `YYYYMMDD` or any automatically recognizable format
        * Response:
            * array of objects, each: (int) planet_id => (bool) is_retrograde
    1. - [ ] If there is no data locally, request the external API, store the data locally and then return the data.
        * Add a `TODO` comment to store data in background.
    1. - [ ] Consider edge cases:
        * Invalid date.
        * No data in the local storage, external APIs are unavailable (no response or time out).
    1. - [ ] Cover `on-date` endpoint with autotests.
        1. - [ ] Mock external API(s).
        1. - [ ] Try to explicitly forbid external connections - if you add a new external API and forget to update
             autotests, the latter will fail while trying to request an external IP.
        1. - [ ] Cover cases: data from external API, data from the local storage, invalid date, no data available.
1. - [ ] Upgrade [luglog.go](../src/general/luglog/luglog.go) calls:
     1. - [ ] Implement logging levels.
     1. - [ ] Add an option to additionally (like `tee`) or exclusively log to files.

## Misc

1. - [ ] Make HTTP server and DBMS accept secured requests (HTTPS + sslmode=require).
1. - [ ] Make a way in tests to specify just then ending part of an endpoint URI (like `/` instead of `/api/`),
     (?) but try using a global router at the same time.
1. - [ ] Cover [api.go](../src/api/api.go) functions with tests.
    1. - [ ] Stripping URI prefixes from API endpoints' URIs.
    1. - [ ] Edge cases with errors.
