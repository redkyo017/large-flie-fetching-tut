# Food Ordering API Server

This project implements a backend API server for a food ordering system using Go and the Fiber framework. It includes modules for managing products, orders, and a robust promo code validation system capable of handling large datasets.

## Table of Contents

* [Features](#features)
* [Project Structure](#project-structure)
* [Getting Started](#getting-started)
    * [Prerequisites](#prerequisites)
    * [Local Setup](#local-setup)
    * [Environment Variables](#environment-variables)
    * [Running the Application](#running-the-application)
* [API Endpoints](#api-endpoints)
    * [OpenAPI Documentation](#openapi-documentation)
    * [Promo Code Validation](#promo-code-validation)
    * [Products](#products)
    * [Orders](#orders)
    * [Debug Endpoints](#debug-endpoints)
* [Running Tests](#running-tests)
* [Promo Code Data Handling](#promo-code-data-handling)

## Features

* **API Implementation:** Comprehensive API for product listing, order creation, and promo code validation.
* **Promo Code Validation:** Validates promo codes based on length (8-10 characters) and presence in at least two source files.
* **Efficient Large File Processing:** Utilizes streaming decompression to temporary disk files and batched aggregation for promo codes, minimizing memory footprint during initial load.
* **Flexible Data Storage:** Supports in-memory storage for promo codes (for development/smaller datasets) and can be switched to PostgreSQL for production-scale data.
* **Clean Architecture:** Structured using `cmd/`, `pkg/`, and `internal/` for clear separation of concerns, maintainability, and scalability.
* **Fiber Framework:** High-performance HTTP server built with Fiber.
* **Graceful Shutdown:** Ensures proper cleanup on application termination.

## Project Structure

## Getting Started

### Prerequisites

* **Go:** Version 1.16 or higher.
* **Git:** For cloning the repository.
* **PostgreSQL (Optional but Recommended for Production Scale):** If you plan to use the database backend for promo codes.

### Local Setup

1.  **Clone the repository:**
    ```bash
    git clone <repository_url> kart-challenge
    cd kart-challenge
    ```
2.  **Initialize Go modules:**
    ```bash
    go mod tidy
    ```
3.  **Create necessary directories:**
    ```bash
    mkdir -p public local_coupons
    ```
4.  **Download OpenAPI Specification Files:**
    The API documentation is served statically. Download these files and place them in the `public/` directory:
    ```bash
    curl -o public/openapi.html https://orderfoodonline.deno.dev/public/openapi.html
    curl -o public/openapi.yaml https://orderfoodonline.deno.dev/public/openapi.yaml
    ```
5.  **Download Local Coupon Files (Optional - for faster development setup):**
    To avoid slow remote downloads during development, you can pre-download the coupon `.gz` files and place them in the `local_coupons/` directory:
    ```bash
    curl -o local_coupons/couponbase1.gz https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz
    curl -o local_coupons/couponbase2.gz https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz
    curl -o local_coupons/couponbase3.gz https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz
    ```

### Environment Variables

Create a `.env` file in the project root (`kart-challenge/.env`) and set the following variables. These will be loaded by the application.

```env
# Application environment: "development" or "production"
APP_ENV=development

# Port for the Fiber server to listen on
PORT=8080

# Path to local .gz coupon files (if APP_ENV is development)
LOCAL_COUPON_DIR=./local_coupons

# Maximum size for decompressed temporary files in MB (e.g., 1024 for 1GB)
MAX_FILE_SIZE_MB=1024

# Set to 'true' to use PostgreSQL for promo code storage, 'false' for in-memory
USE_DATABASE=false

# PostgreSQL connection string (only required if USE_DATABASE is true)
# Example: postgres://user:password@localhost:5432/mydatabase?sslmode=disable
# Make sure to replace user, password, host, port, and dbname with your actual PostgreSQL credentials.
DATABASE_URL="postgres://your_user:your_password@localhost:5432/your_database_name?sslmode=disable"