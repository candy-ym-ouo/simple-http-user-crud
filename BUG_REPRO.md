# Bug Reproduction

## Baseline

This reproduction is based on branch `base_bug_005`.

## Steps

1. Submit `POST /users` with a valid name and a display-name email such as `"张三" <zhangsan@example.com>`.
2. Submit `PUT /users/1` with the same kind of email input after creating a valid user.

## Actual Result

Both requests can be accepted because the standard-library email parser accepts RFC 5322 display names, angle brackets, and comments in addition to a plain mailbox address.

## Expected Result

The API accepts only a plain email address, such as `zhangsan@example.com`. Any display name, angle brackets, or comment must receive the existing invalid-email response without changing stored data.

## Focused Check

```bash
go test ./internal/handler -run TestRejectDisplayNameEmail -count=1
```

On the `base_bug_005` baseline, display-name email inputs are incorrectly accepted on the create and update paths.
