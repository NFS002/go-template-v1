# nfs002/template/v1 -  API Token auth template

A relatively simple golang application template with the following features:

- An HTTP server using JSON over REST
- API token authentication with scoped tokens
- Each endpoint can be configured to require a given scope
- A user can request a token with a given scope with their username/email and password
    - The token is only granted if the user's scope has at least all of the requested scope
- When the app starts, it looks for an environemnt variable `APP_ENV`
    - If this is set, it attempts to load the environment variables in the file `.env.${APP_ENV}`
        - If file does not exist, it attempts to load the environment variables in the default file `.env`
    - If this is not set, it attempts to load the environment variables in the default file `.env`
- Connection to the db instance from the app is set via the `$POSTGRESQL_URL` environment variable
- Database migrtations managed by the [golang-migrate](https://github.com/golang-migrate/migrate) module
    - All up migrations are run on startup if `RUN_MIGRAGTIONS=true`
- Tokens and users persisted in a postgresql database
    - Passwords are hashed and salted using bcrypt before persisting
    - Tokens are hashed using SHA-256 before persisting
- Request validation using the [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) module
- Postgresql db intstance runs in a docker container
    - start with `docker-compose up`

## The following table lists all API endpoints, their behavior, and their required token scope:

<details>
<summary>View table</summary>

| Route                          | METHOD | Description                                                 | Authenticated      | Scope                             |
| ------------------------------ | ------ | ----------------------------------------------------------- | ------------------ | --------------------------------- |
| /hello                         | GET    | Say a generic hello                                         | No                 | none                              |
| /api/authenticate              | POST   | Returns a token for the given user with the requested scope | With user password | none                              |
| /api/read-a/hello-user         | GET    | Say hello to the calling user (associated with the token)   | Yes                | read:a                            |
| /api/read-a-write-a/hello-user | GET    | Say hello to the calling user (associated with the token)   | Yes                | read:a, write:a                   |
| /api/admin/hello-user          | GET    | Say hello to the calling user (associated with the token)   | Yes                | read:a, write:a, read:b, write: b |
| /api/admin/users               | GET    | Get all registered users                                    | Yes                | read:a, write:a, read:b, write: b |
| /api/admin/users               | POST   | Create a new user                                           | Yes                | read:a, write:a, read:b, write: b |
| /api/admin/users/:userId       | GET    | Get the user with the given userId                          | Yes                | read:a, write:a, read:b, write: b |
| /api/admin/users/:userId       | PUT    | Update the user with the given userId                       | Yes                | read:a, write:a, read:b, write: b |
| /api/admin/users/:userId       | DELETE | Delete the user with the given userId                       | Yes                | read:a, write:a, read:b, write: b |

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

### TODO:
1. Token expiry
2. Allow a user to have multiple tokens
2. Order of env loading
3. Logging and middleware
4. Remove unecessary stripe stuff
5. APP_ENV and config env
6. Test update and delete endpoints
7. Add basic unit tests
8. Env file comments
9. Add more docs to README
10. Auth middleware applied to group not individual routes. Use gin ?
