# Bug Reproduction

## Baseline

This reproduction is based on branch `base_bug_004`.

## Steps

1. Create a user with `POST /users` and retain its returned ID (`1` in the test case).
2. Update that user with `PUT /users/1`. The request JSON intentionally contains no `id` field.
3. Inspect the update response, then fetch the same user with `GET /users/1`.

## Actual Result

The update path stores the request-body model unchanged. Because the decoded body has an ID of zero, the saved user and the update response expose `"id":0` rather than the ID selected by the URL.

## Expected Result

`PUT /users/1` must preserve ID `1` in both the persisted user and the response, so later requests can continue to address that user by the same ID.

## Focused Check

```bash
go test ./internal/handler -run TestUserCRUD -count=1
```

On the `base_bug_004` baseline, the ID-preservation assertion fails after the update request.
