# Go Todo Demo

A feature-incomplete, not-so-perfect, not-blazingly-fast and not-production-ready todo api I wrote while exploring some of the features in Go :laughing:

## Running the app

-   With in-memory storage (`map[int]string` :smile:)

    ```sh
    go run .
    ```

-   With MongoDB docker container (refer `docker-compose.yaml`)

    ```sh
    docker compose up

    go run . -m true
    ```

    (Note: use `Ctrl+C` to stop the app. And, if docker container was used, run `docker compose down` to stop the container)

## APIs

-   **Note**:

    -   `localhost:3000` is used as the host & port in the examples)
    -   `jq` is used in the examples to pretty-print json response
    -   use a trailing slash when sending the request using curl i.e. `/todo/`

-   **Add** a Todo

    ```sh
    curl -X POST localhost:3000/todo/ \
        -H 'Content-Type: application/json' \
        -d '{ "description": "test_description", "due": "9999-12-01 11:59:59PM" }' | jq .
    ```

-   **Delete** a Todo

    ```sh
    curl localhost:3000 | jq .
    ```

-   **Get** a Todo

    ```sh
    curl localhost:3000/todo/1 | jq .
    ```

-   **Get All** Todos

    ```sh
    curl localhost:3000/todo/ | jq .
    ```
