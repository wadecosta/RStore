# RStore

---

RStore is a custom in-house inventory and store management application designed to help manage the operations of a small retail business. This project was created as a practical software engineering exercise while exploring backend development, database design, security practices, and containerized deployment.

## Features

### Currently Implemented

- User authentication and session management
- Secure password hashing using modern cryptographic practices
- User profiles and account management
- Vendor management system
- MariaDB database integration
- PDF generation and document handling
- Responsive frontend using Bootstrap
- Containerized deployment using Docker/Podman

## Technologies Used

### Backend

- Go
- MariaDB
- SCS session management
- Go standard library HTTP server
- pdfcpu for PDF processing

### Frontend

- HTML
- CSS
- JavaScript
- Bootstrap 5
- Bootstrap Icons

### Deployment

- Docker / Podman
- Docker Compose / Podman Compose
- Alpine Linux containers

## Running Locally
``git clone https://github.com/wadecosta/RStore.git
cd RStore
go mod download
podman-compose up --build``


- The Application will start on `http://localhost:8081`
