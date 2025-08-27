# ✨ React-Golang Starter Kit ✨

This project serves as a robust and modern starter kit for building full-stack applications, seamlessly integrating a React frontend with a high-performance Golang backend. Designed for rapid development and scalability, it provides a solid foundation with best practices already in place.

## 🚀 Features

- **⚛️ React Frontend:**
  - Built with [Vite](https://vitejs.dev/) for blazing-fast development.
  - [React Router](https://reactrouter.com/en/main) for declarative navigation.
  - [TailwindCSS](https://tailwindcss.com/) for utility-first styling.
  - [ShadCN UI](https://ui.shadcn.com/) components for a beautiful and accessible user interface.
  - **[Vitest](https://vitest.dev/)** for fast unit and component testing.
  - Optimized for performance and developer experience.
- **⚙️ Golang Backend:**
  - Powered by the [Fiber Web Framework](https://gofiber.io/) for a fast and flexible API.
  - [Air](https://github.com/cosmtrek/air) for live reloading during development.
  - [GORM](https://gorm.io/) for elegant Object-Relational Mapping (ORM) with PostgreSQL.
  - Structured project layout for maintainability and scalability.
  - Includes basic CRUD operations and authentication scaffolding.
- **🐳 Docker Support:**
  - `Dockerfiles` for both frontend and backend for easy containerization.
  - Simplified deployment and consistent development environments.
- **💾 Database Integration:**
  - Pre-configured for PostgreSQL, allowing quick setup and integration.
  - Scalable and reliable data storage solution.
- **✅ API Testing:**
  - Integrated tools for efficient API testing to ensure robustness.

## 🏁 Getting Started

Follow these steps to get your development environment up and running.

### Prerequisites

Ensure you have the following installed on your system:

- **Git:** For version control.
- **Node.js (LTS) & npm/yarn:** For frontend development.
- **Go (1.20+):** For backend development.
- **Docker & Docker Compose (Optional):** Highly recommended for isolated development environments and deployment.
- **PostgreSQL:** Database server.

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
    cd react-golang-starter-kit
    ```

2.  **Set up Environment Variables:**
    Create a `.env` file in the project root based on `.env.example`.

    ```bash
    # Backend Environment Variables
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=user
    DB_PASSWORD=password
    DB_NAME=mydatabase
    JWT_SECRET=supersecretkey
    API_PORT=8080

    # Frontend Environment Variables
    VITE_API_URL=http://localhost:8080/api
    ```

3.  **Backend Setup:**

    ```bash
    cd backend
    go mod tidy          # Download dependencies
    go install github.com/cosmtrek/air@latest  # Install Air for live reloading
    air                  # Start the backend server with live reloading
    ```

    Alternatively, you can run without Air:

    ```bash
    go run cmd/main.go   # Start the backend server
    ```

4.  **Frontend Setup:**

    ```bash
    cd ../frontend
    npm install          # Install frontend dependencies
    npm run dev          # Start the frontend development server
    ```

Your application should now be running!

## 🧪 Testing

### Frontend (React with Vitest)

To run the frontend tests, navigate to the `frontend` directory and use the following command:

```bash
cd frontend
npm test
```

This will execute all tests defined using Vitest.

## 📂 Project Structure

```
react_golang_starter_kit/
├── backend/                  # 🚀 Golang Backend
│   ├── cmd/                  # Application entry point
│   ├── internal/             # Internal packages (handlers, models, database)
│   ├── pkg/                  # Reusable packages (if any)
│   ├── Dockerfile            # Dockerfile for backend
│   ├── go.mod                # Go module definition
│   └── ...
├── frontend/                 # 🌐 React Frontend
│   ├── public/               # Static assets
│   ├── app/                  # Main application source code
│   │   ├── components/       # Reusable React components
│   │   ├── hooks/            # Custom React hooks
│   │   ├── lib/              # Utility functions
│   │   ├── routes/           # React Router routes
│   │   └── types/            # TypeScript type definitions
│   ├── Dockerfile            # Dockerfile for frontend
│   ├── package.json          # Node.js package configuration
│   └── ...
├── documentations/           # 📚 Project Documentation
│   └── starter_kit/          # Specific documentation for this starter kit
│   └── third_party/          # Documentation for third-party tools/libraries
└── README.md                 # 📄 Project Overview and Setup Instructions
```

## 🔒 Environment Variables

Critical environment variables are managed through `.env` files. A `.env.example` is provided for reference. It is crucial to set these values correctly for the application to function.

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL database connection details.
- `JWT_SECRET`: Secret key for JWT authentication. **(Highly recommended to change in production!)**
- `API_PORT`: Port on which the Golang backend API will run.
- `VITE_API_URL`: Frontend URL to access the backend API.

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1.  Fork the repository.
2.  Create a new branch (`git checkout -b feature/your-feature-name`).
3.  Make your changes.
4.  Commit your changes (`git commit -m 'feat: Add new feature'`).
5.  Push to the branch (`git push origin feature/your-feature-name`).
6.  Open a Pull Request.

Please ensure your code adheres to the existing style and conventions.

## 📄 License

This project is licensed under the MIT License - see the `LICENSE` file for details.
