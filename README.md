
# Conso-CoinHub-DataCollector

Conso-CoinHub-DataCollector is a data collector application designed to interact with the OKX Web3 API. It efficiently gathers data and stores it in a PostgreSQL database while utilizing Redis for temporary data storage.

## Project Structure

```
Conso-CoinHub-DataCollector
├── cmd
│   └── main.go          # Entry point of the application
├── internal
│   ├── api
│   │   └── okx.go      # Functions to interact with the OKX Web3 API
│   ├── db
│   │   └── postgres.go  # Manages PostgreSQL database connection
│   ├── cache
│   │   └── redis.go     # Manages Redis cache connection
│   ├── collector
│   │   └── collector.go  # Implements data collection logic
│   └── config
│       └── config.go    # Handles application configuration settings
├── go.mod                # Module definition and dependencies
├── go.sum                # Checksums for module dependencies
└── README.md             # Project documentation
```

## Setup Instructions

1. **Clone the repository:**
   ```
   git clone https://github.com/yourusername/Conso-CoinHub-DataCollector.git
   cd Conso-CoinHub-DataCollector
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Configure environment variables:**
   Create a `.env` file in the root directory and set the necessary environment variables for PostgreSQL and Redis connections.

4. **Run the application:**
   ```
   go run cmd/main.go
   ```

## Usage

Once the application is running, it will start collecting data from the OKX Web3 API and store it in the configured PostgreSQL database. Temporary data will be cached in Redis for quick access.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.