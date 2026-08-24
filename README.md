# RS-GoAPIServer

This project is a foundational Go-based backend API server designed for extensibility and robustness. It provides a solid starting point for building scalable and maintainable web services.

## Core Architecture

The server is built with a modular architecture, separating concerns into distinct packages. This design makes it easier to understand, maintain, and extend the codebase.

### Key Packages

-   **`mainserver`**: The entry point of the application. It initializes the server, database connections, and other core components.
-   **`database`**: Manages the database connection pool and provides a `sqlx.DB` object for database interactions.
-   **`entities`**: Defines the data models (structs) that represent the core business objects of the application.
-   **`approutes` & `adminroutes`**: These packages define the API endpoints for the application and admin functionalities, respectively.
-   **`authentication`**: Handles user authentication, including token generation and validation.
-   **`middleware`**: Contains HTTP middleware for tasks like logging, authentication, and panic recovery.
-   **`apierrors`**: Provides a centralized error handling mechanism.
-   **`global`**: Contains global variables and configuration, such as environment variables.

## Getting Started

### Prerequisites

-   Go 1.20 or higher
-   MySQL database

### Setup

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd rs-goapiserver
    ```

2.  **Configuration:**
    -   Copy `env.json.example` to `env.json`.
    -   Update `env.json` with your database credentials and other environment-specific settings.

3.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```

4.  **Run the server:**
    ```bash
    go run .
    ```

The server will start on the port specified in your `env.json` file (default is `9990`).

## How It Works

### Server Initialization

The server startup process is handled in the `mainserver` package. Here's a summary of the steps:

1.  **Context Management**: A root context is created to manage the server's lifecycle, allowing for graceful shutdowns.
2.  **Database Connection**: The `database.StartDB` function is called to establish a connection to the MySQL database.
3.  **HTTP Server Configuration**: An `http.Server` is configured with timeouts and a custom `ServeMux`.
4.  **Panic Recovery**: A middleware is set up to recover from panics, log the error, and return a generic error response to the client.
5.  **Graceful Shutdown**: The server listens for interrupt signals (`os.Interrupt`, `syscall.SIGTERM`) to shut down gracefully, allowing in-flight requests to complete.

### Routing

API routes are defined in the `approutes` and `adminroutes` packages. The server uses the standard `http.ServeMux` for routing. To add a new route, you can add a new handler function in the appropriate routes package.

### Database Interaction

The project uses `sqlx` for database interactions, which simplifies common database operations and provides a convenient way to map database rows to Go structs. The `database.DB` object is a global `sqlx.DB` instance that can be used throughout the application.

### Authentication

Authentication is handled by the `authentication` package. It provides functions for password hashing, API key generation, and user authentication.

### Error Handling

The `apierrors` package provides a centralized `HandleError` function for logging and responding to errors. This ensures consistent error handling across the application.

## Extending the Server

### Adding a New Entity and API Endpoint

1.  **Create a new entity struct** in the `entities` package (e.g., `entities/product/product.go`).
2.  **Create a new routes file** in the `approutes` package (e.g., `approutes/product_routes.go`).
3.  **Define your handler functions** in the new routes file. These functions will contain the business logic for your new entity.
4.  **Register your new routes** in the `mainserver` package.
5.  **Use the `database.DB` object** to interact with the database.

This structure allows for a clean separation of concerns and makes it easy to add new features to the server.
