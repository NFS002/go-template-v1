# nfs002/template/v1 -  API Token auth template

A relatively simple golang application template with the following features:

- An HTTP server using JSON over REST
- API token authentication with scoped tokens
- Each endpoint can be configured to require a given scope
- A user can request a token with a given scope with their username/email and password
    - The token is only granted if the user's scope has at least all of the requested scope
- Lazy expired token cleanup
    - When a request is sent using an expired token, the token is deleted from the database
- When the app starts, it looks for an environemnt variable `APP_ENV`
    - `$APP_ENV` must be set manually before running the app
        - e.g `APP_ENV=dev go run .`
    - If this is set, it attempts to load the environment variables in the file `.env.${APP_ENV}`
        - If file does not exist, it attempts to load the environment variables in the default file `.env`
    - If this is not set, it attempts to load the environment variables in the default file `.env`
    - Connection to the db instance from the app is set via the `$POSTGRESQL_URL` environment variable
    - Uses the [godotenv](https://github.com/joho/godotenv) module for loading env files
- Database migrtations managed by the [golang-migrate](https://github.com/golang-migrate/migrate) module
    - All up migrations are run on startup if `RUN_MIGRAGTIONS=true`
- Uses the [zerolog](https://github.com/rs/zerolog) module for logging
- Tokens and users persisted in a postgresql database
    - Passwords are hashed and salted using bcrypt before persisting
    - Tokens are hashed using SHA-256 before persisting
- Request validation using the [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) module
- Postgresql db intstance runs in a docker container
    - start with `docker-compose up`

## Running the app
```sh
APP_ENV=dev go run . # Will load environment variables from a file named .env.dev
```

or

```sh
APP_ENV=dev go run . # Will load environment variables from the a file named .env
```

## DB Schema

- Table: users
    - id
    - first_name
    - last_name
    - email
    - password (brcrypt hash)
    - scope (the maximum scope a user can reques an auth token for)
    - updated_at
    - created_at

- Table: tokens
    - id
    - user_id (foreign key constraint references users.id, cascade delete)
    - token_hash (SHA-256 Hash)
    - expiry
    - scope
    - updated_at
    - created_at

- A trigger also exists on both tables to automatically set `updated_at` on a row to the current time whenever a row is updated.



## The following table lists all API endpoints, their behavior, and their required token scope:

<details>
<summary>View table</summary>

| Route                          | METHOD | Description                                                 | Authenticated      | Scope                             |
| ------------------------------ | ------ | ----------------------------------------------------------- | ------------------ | --------------------------------- |
| /hello                         | GET    | Say a generic hello                                         | No                 | none                              |
| /api/authenticate              | POST   | Returns a token for the given user with the requested scope | With user password | none                              |
| /api/hello-user                | GET    | Say hello to the calling user (associated with the token)   | Bearer Token       | none                              |
| /api/read-a/hello-user         | GET    | Say hello to the calling user (associated with the token)   | Bearer Token       | read:a                            |
| /api/read-a-write-a/hello-user | GET    | Say hello to the calling user (associated with the token)   | Bearer Token       | read:a, write:a                   |
| /api/admin/hello-user          | GET    | Say hello to the calling user (associated with the token)   | Bearer Token       | read:a, write:a, read:b, write: b |
| /api/admin/users               | GET    | Get all registered users                                    | Bearer Token       | read:a, write:a, read:b, write: b |
| /api/admin/users               | POST   | Create a new user                                           | Bearer Token       | read:a, write:a, read:b, write: b |
| /api/admin/users/:userId       | GET    | Get the user with the given userId                          | Bearer Token       | read:a, write:a, read:b, write: b |
| /api/admin/users/:userId       | PUT    | Update the user with the given userId                       | Bearer Token       | read:a, write:a, read:b, write: b |
| /api/admin/users/:userId       | DELETE | Delete the user with the given userId                       | Bearer Token       | read:a, write:a, read:b, write: b |

*These endpoints and their scopes have no meaning... they are configured like this **purely** for demonstration/testing*
</details>

## The database migration [seed_users_table](migrations/000002_seed_users_table.up.sql) adds the following set of test users

<details>
<summary> Test users </summary>

| **first_name** | **last_name** | **email**         | **password** | **scope**                     |
|----------------|---------------|-------------------|--------------|-------------------------------|
| User           | One           | user@example.com  | secret       | read:a,write:a,read:b,write:b |
| User           | Two           | user2@example.com | secret       | read:a,read:b                 |
| User           | Three         | user3@example.com | secret       | write:b                       |


*Passwords are hashed and salted using bcrypt, so the above is **not** a database representation*

</details>

## Technologies/requirements
- Golang
- Docker/Docker-compose
    - Official postgresql image (v16.2)
- Golang migrate CLI (optional)

*This is just an application template, to be extended/forked if it contains some of your requirements, it is not intended to have any use case aside from this*
