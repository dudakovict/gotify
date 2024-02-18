# Gotify

[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/dudakovict/gotify)](https://github.com/dudakovict/gotify)

Gotify is an application that provides authorization/authentication, user creation, and a notification system. 
Users can subscribe to topics and receive email notifications when there are updates.

## Implementation Notes

Distribution and processing of tasks is done via a asynchronous *Redis* worker
and is run side-by-side with our API in a separate goroutine. Both goroutines
are managed by an errgroup and support graceful shutdown and load shedding. For
the data layer I decided to use *sqlx* in combination with the *pgx* driver.

## Running locally

```
$ docker-compose up --build
```

## Docs

Docs are served from the API [here](http://localhost:3000/swagger/index.html) when the server is running.

## Features
- **Authorization/Authentication**: Secure register & login system to protect users accounts.
- **Notification System**: Users can subscribe to topics and receive email notifications.
- **Retry Mechanism**: Sending notifications is being retried upon failure.
- **Email Verification**: Users are sent a verification email when registered.