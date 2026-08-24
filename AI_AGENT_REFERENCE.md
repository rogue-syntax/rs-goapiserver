# AI Agent Reference for RS-GoAPIServer

This document provides a structured overview of the `rs-goapiserver` project, intended for use by AI agents to understand the codebase and assist in development.

## Project Overview

-   **Name**: `rs-goapiserver`
-   **Language**: Go
-   **Description**: An extendable base for a backend API server.
-   **Main Entrypoint**: The `main` function is expected to be in the root of the project, which then calls `mainserver.Serve`.

## Core Components

| Package             | Responsibility                                                              | Key Files/Functions                               |
| ------------------- | --------------------------------------------------------------------------- | ------------------------------------------------- |
| `mainserver`        | Initializes and runs the HTTP server, database, and other core services.    | `Serve()`                                         |
| `database`          | Manages the database connection pool.                                       | `StartDB()`, `DB` (global `*sqlx.DB` object)      |
| `entities`          | Defines data models (structs) for the application.                          | `entities/user/user.go`                           |
| `approutes`         | Defines public-facing API routes.                                           | `approutes.go`                                    |
| `adminroutes`       | Defines administrative API routes.                                          | `adminroutes.go`                                  |
| `authentication`    | Handles user authentication, password hashing, and API key management.      | `authentication.go`                               |
-   **`middleware`**: Contains HTTP middleware for logging, authentication, and panic recovery.
-   **`apierrors`**: Provides a centralized error handling mechanism (`HandleError` function).
-   **`global`**: Contains global configuration and environment variables.

## Development Workflow

### Adding a New Entity and API Endpoint

To add a new feature (e.g., a "product" entity), follow these steps:

1.  **Define the Entity**:
    -   Create a new package under `entities/`, for example, `entities/product/`.
    -   In that package, create a file (e.g., `product.go`) and define the `Product` struct.

2.  **Create Route Handlers**:
    -   Create a new file in the `approutes/` package (e.g., `product_routes.go`).
    -   Implement the HTTP handler functions for the product-related endpoints (e.g., `createProduct`, `getProduct`, `updateProduct`, `deleteProduct`).
    -   These handlers will use the `database.DB` object for database operations.

3.  **Register Routes**:
    -   In `mainserver/mainserver.go`, import the new routes package.
    -   In the `Serve` function, register the new handlers with the `http.ServeMux`.

    ```go
    // Example in mainserver/mainserver.go
    // ...
    import "github.com/rogue-syntax/rs-goapiserver/approutes"
    // ...
    func Serve(mux *http.ServeMux, root context.Context) (context.Context, *http.ServeMux) {
        // ...
        mux.HandleFunc("/v1/products", approutes.CreateProductHandler)
        // ...
    }
    ```

## Key Concepts

-   **Database Access**: Use the global `database.DB` object (`*sqlx.DB`) for all database queries.
-   **Error Handling**: Use the `apierrors.HandleError` function for consistent error logging and response formatting.
-   **Configuration**: Environment-specific configuration is loaded from `env.json` into the `global.EnvVars` struct.
-   **Authentication**: The `authentication` package provides utilities for securing endpoints. Middleware should be used to protect routes that require authentication.
-   **Graceful Shutdown**: The server is designed to shut down gracefully, allowing active requests to complete. This is managed by the root context in `mainserver`.
