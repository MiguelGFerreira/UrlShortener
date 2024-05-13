## URL Shortener - Go Application

This project implements a simple URL shortener service in Go. It allows users to shorten long URLs and provides redirection to the original URL when the shortened alias is accessed.

### Features

* Shortens long URLs using a database for storage.
* Redirects users to the original URL when accessing the shortened alias.

### Prerequisites

* Go installed (version 1.18 or later recommended)
* PostgreSQL database server
* Text editor or IDE


### Installation

1. **Clone the repository:**

```bash
git clone https://github.com/your-username/UrlShortener.git
```

2. **Install dependencies:**

```bash
go mod download
```

3. **Configure database credentials:**

   * Set the following environment variables with your PostgreSQL credentials:
      * `DB_USER`: Username for your PostgreSQL database.
      * `DB_PASS`: Password for your PostgreSQL database.

4. **Run the services:**

   * **Shortener Service:**
     ```bash
     go run shortener/main.go
     ```
     This starts the service that listens for URL shortening requests on port 8080 (default).

   * **Redirector Service:**
     ```bash
     go run redirecionador/main.go
     ```
     This starts the service that handles redirection requests for shortened URLs on port 8081 (default).


### Usage

1. **Shortening URLs:**

   * Open a web browser and navigate to `http://localhost:8080`.
   * Enter a long URL in the text field and click "Shorten".
   * The service will generate a shortened alias for the URL and display it.

2. **Accessing Shortened URLs:**

   * Copy the shortened alias generated in the previous step.
   * Paste the shortened alias into your web browser's address bar and press Enter.
   * You will be redirected to the original long URL.


### Notes

* This is a simple example and can be extended to include features like:
    * Custom URL aliases.
    * Usage statistics for shortened URLs.
    * Security measures to prevent abuse.
* Remember to replace `your-username` in the clone command with your actual GitHub username.


### Contributing

Feel free to fork the repository and submit pull requests with your improvements or bug fixes.
