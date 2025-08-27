package main

import (
	"log"
	"net/http"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

const swaggerJSON = `{
    "swagger": "2.0",
    "info": {
        "contact": {},
        "title": "React Go Starter Kit API",
        "description": "API documentation for the React Go Starter Kit",
        "version": "1.0.0"
    },
    "host": "localhost:8080",
    "basePath": "/api",
    "schemes": ["http"],
    "paths": {
        "/health": {
            "get": {
                "description": "Get the health status of the server",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["health"],
                "summary": "Check server health status",
                "responses": {
                    "200": {
                        "description": "status: ok, message: Server is running",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/users": {
            "get": {
                "description": "Retrieve a list of all users",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Get all users",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.User"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new user with the provided information",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Create a new user",
                "parameters": [
                    {
                        "description": "User object",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    },
                    "400": {
                        "description": "Invalid JSON",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to create user",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users/{id}": {
            "get": {
                "description": "Retrieve a single user by their ID",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Get a user by ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    },
                    "400": {
                        "description": "Invalid user ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "put": {
                "description": "Update a user's information by their ID",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Update an existing user",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated user object",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    },
                    "400": {
                        "description": "Invalid user ID or JSON",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to update user",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete a user by their ID",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Delete a user",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid user ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to delete user",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.User": {
            "type": "object",
            "required": ["email", "name"],
            "properties": {
                "id": {
                    "description": "The unique ID of the user",
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "description": "The name of the user",
                    "type": "string",
                    "example": "John Doe"
                },
                "email": {
                    "description": "The email address of the user (must be unique)",
                    "type": "string",
                    "example": "john.doe@example.com"
                },
                "created_at": {
                    "description": "When the user was created",
                    "type": "string",
                    "example": "2023-08-27T12:00:00Z"
                },
                "updated_at": {
                    "description": "When the user was last updated",
                    "type": "string",
                    "example": "2023-08-27T12:00:00Z"
                }
            }
        }
    }
}`

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	database.ConnectDB()

	// Create Chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware for React frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080"}, // React dev server + backend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	setupRoutes(r)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setupRoutes(r chi.Router) {
	// Simple test route at root level
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test route working!"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API test route working!"))
		})
		r.Get("/health", handlers.HealthCheck)

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", handlers.GetUsers)       // GET /api/users
			r.Post("/", handlers.CreateUser)     // POST /api/users
			r.Get("/{id}", handlers.GetUser)     // GET /api/users/{id}
			r.Put("/{id}", handlers.UpdateUser)  // PUT /api/users/{id}
			r.Delete("/{id}", handlers.DeleteUser) // DELETE /api/users/{id}
		})
	})

	// Swagger routes
	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/index.html")
	})
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})
}
