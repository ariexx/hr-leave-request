# Leave HR Requests


## How to setup the application :
1. Install Go Version 1.25
2. Clone this repository using the command:
   ```
   git clone
    ```
3. Navigate to the project directory:
4. Install the required dependencies using the command:
   ```
   go mod tidy
   ```
5. Copy the `config.yaml.example` file to `config.yaml` and update the configuration values as needed.
6. Run the database migrations using the command:
   ```
   migrate -path internal/migrations -database "mysql://root:@tcp(localhost:3306)/hr_leave_requests" up
   ```
   change the database connection string as per your setup.
7. Start the application using the command:
   ```
   go run main.go
   ```

## Overview the Application :
This application is designed to manage leave requests for employees in an organization. It provides functionalities for employees to submit leave requests, view their leave history, and for HR personnel to approve or reject these requests.
### Key Features:
- Employee Registration and Authentication
- Submit Leave Requests
- View Leave History
- Approve or Reject Leave Requests (for HR)

### Technologies Used:
- Go (Golang) for backend development
- MySQL for database management
- YAML for configuration management

### Project Structure:
- `main.go`: The entry point of the application.
- `config.yaml`: Configuration file for the application settings.
- `internal/migrations/`: Contains database migration files.
- `go.mod` and `go.sum`: Go module files for dependency management.
- `README.md`: Documentation for the application.
- `config/` : Contains configuration related code and files.
- `handlers/` : Contains HTTP handler functions for various endpoints.
- `models/` : Contains data models and database interaction logic.
- `injector/` : Contains dependency injection related code.
- `services/` : Contains business logic and service layer code.
- `repositories/` : Contains data access layer code.
- `middleware/` : Contains middleware functions for request processing.
- `routes/` : Contains route definitions and setup.

### How to run tests :
1. Ensure you have a testing database set up.
2. Run the tests using the command:
   ```
   go test ./...
   ```
   
### How to run with docker :
1. Ensure you have Docker installed on your machine.
2. Build the Docker image and run the container using the command:
   ```
   docker-compose up --build -d
   ```
3. Access the application at `http://localhost:9090` (or the port specified in your `docker-compose.yml` file).