import { useState, useEffect } from "react";
import { Link } from "react-router";
import { toast } from "sonner";
import { API_BASE_URL } from "../lib/api";

interface User {
  id: number;
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string | null;
  name: string;
  email: string;
}

interface HealthResponse {
  status: string;
  message: string;
}

export function Demo() {
  const [healthStatus, setHealthStatus] = useState<HealthResponse | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);

  const [users, setUsers] = useState<User[]>([]);
  const [usersLoading, setUsersLoading] = useState(false);

  const [newUser, setNewUser] = useState({ name: "", email: "" });
  const [createLoading, setCreateLoading] = useState(false);

  const API_BASE = `${API_BASE_URL}/api`;

  // Test health check
  const testHealthCheck = async () => {
    setHealthLoading(true);
    try {
      const response = await fetch(`${API_BASE}/health`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data: HealthResponse = await response.json();
      setHealthStatus(data);
      toast.success("Health check successful!", {
        description: `Status: ${data.status} - ${data.message}`,
      });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      toast.error("Health check failed", {
        description: errorMessage,
      });
    } finally {
      setHealthLoading(false);
    }
  };

  // Fetch all users
  const fetchUsers = async () => {
    setUsersLoading(true);
    try {
      const response = await fetch(`${API_BASE}/users`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data: User[] = await response.json();
      setUsers(data);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      toast.error("Failed to fetch users", {
        description: errorMessage,
      });
    } finally {
      setUsersLoading(false);
    }
  };

  // Create a new user
  const createUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newUser.name.trim() || !newUser.email.trim()) {
      toast.error("Validation Error", {
        description: "Please fill in all fields",
      });
      return;
    }

    setCreateLoading(true);

    try {
      const response = await fetch(`${API_BASE}/users`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(newUser),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const createdUser: User = await response.json();
      toast.success("User created successfully!", {
        description: `Welcome ${createdUser.name}!`,
      });
      setNewUser({ name: "", email: "" });
      // Refresh the users list
      fetchUsers();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      toast.error("Failed to create user", {
        description: errorMessage,
      });
    } finally {
      setCreateLoading(false);
    }
  };

  // Delete a user
  const deleteUser = async (
    userId: number,
    userName: string,
  ): Promise<void> => {
    if (!confirm(`Are you sure you want to delete user "${userName}"?`)) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE}/users/${userId}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      toast.success("User deleted successfully!", {
        description: `${userName} has been removed from the system.`,
      });
      // Refresh the users list
      fetchUsers();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      toast.error("Failed to delete user", {
        description: errorMessage,
      });
    }
  };

  // Load users on component mount
  useEffect(() => {
    fetchUsers();
  }, []);

  // Add loading toast when initially fetching users
  useEffect(() => {
    if (usersLoading && users.length === 0) {
      toast.loading("Loading users...", {
        id: "fetch-users",
      });
    } else {
      toast.dismiss("fetch-users");
      if (!usersLoading && users.length > 0) {
        toast.success("Users loaded successfully!", {
          description: `Found ${users.length} user${users.length !== 1 ? "s" : ""}`,
        });
      }
    }
  }, [usersLoading, users]);

  return (
    <main className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8 px-4">
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Header */}
        <header className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-2">
            React + Go Integration Test
          </h1>
          <p className="text-gray-600 dark:text-gray-300">
            Test your backend API endpoints from this React frontend
          </p>
        </header>

        {/* Health Check Section */}
        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
            🔍 Health Check
          </h2>
          <div className="space-y-4">
            <button
              onClick={testHealthCheck}
              disabled={healthLoading}
              className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white font-medium py-2 px-4 rounded-lg transition-colors"
            >
              {healthLoading ? "Testing..." : "Test Health Check"}
            </button>

            {healthStatus && (
              <div className="p-3 bg-green-100 dark:bg-green-900 border border-green-300 dark:border-green-700 rounded-lg">
                <p className="text-green-700 dark:text-green-300">
                  ✅ Status: {healthStatus.status}
                </p>
                <p className="text-green-600 dark:text-green-400 text-sm">
                  {healthStatus.message}
                </p>
              </div>
            )}
          </div>
        </section>

        {/* Create User Section */}
        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
            👤 Create User
          </h2>
          <form onSubmit={createUser} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label
                  htmlFor="name"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                  Name
                </label>
                <input
                  type="text"
                  id="name"
                  value={newUser.name}
                  onChange={(e) =>
                    setNewUser({ ...newUser, name: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
                  placeholder="Enter user name"
                  disabled={createLoading}
                />
              </div>
              <div>
                <label
                  htmlFor="email"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                  Email
                </label>
                <input
                  type="email"
                  id="email"
                  value={newUser.email}
                  onChange={(e) =>
                    setNewUser({ ...newUser, email: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
                  placeholder="Enter email address"
                  disabled={createLoading}
                />
              </div>
            </div>
            <button
              type="submit"
              disabled={createLoading}
              className="bg-green-600 hover:bg-green-700 disabled:bg-green-300 text-white font-medium py-2 px-4 rounded-lg transition-colors"
            >
              {createLoading ? "Creating..." : "Create User"}
            </button>
          </form>
        </section>

        {/* Users List Section */}
        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">
              📋 Users List
            </h2>
            <button
              onClick={() => {
                fetchUsers();
                if (users.length > 0) {
                  toast.info("Refreshing users list...");
                }
              }}
              disabled={usersLoading}
              className="bg-gray-600 hover:bg-gray-700 disabled:bg-gray-300 text-white font-medium py-2 px-4 rounded-lg transition-colors"
            >
              {usersLoading ? "Loading..." : "Refresh"}
            </button>
          </div>

          {usersLoading ? (
            <div className="text-center py-8">
              <p className="text-gray-600 dark:text-gray-300">
                Loading users...
              </p>
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400">
                No users found. Create one above!
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-600">
                    <th className="text-left py-2 px-4 text-gray-900 dark:text-white font-medium">
                      ID
                    </th>
                    <th className="text-left py-2 px-4 text-gray-900 dark:text-white font-medium">
                      Name
                    </th>
                    <th className="text-left py-2 px-4 text-gray-900 dark:text-white font-medium">
                      Email
                    </th>
                    <th className="text-left py-2 px-4 text-gray-900 dark:text-white font-medium">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => {
                    const userUrl = `/users/${user.id}`;
                    console.log("Demo - User data:", {
                      id: user.id,
                      name: user.name,
                      email: user.email,
                      generatedUrl: userUrl,
                    });

                    return (
                      <tr
                        key={user.id}
                        className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700"
                      >
                        <td className="py-2 px-4 text-gray-700 dark:text-gray-300">
                          {user.id}
                        </td>
                        <td className="py-2 px-4">
                          <Link
                            to={userUrl}
                            className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline font-medium"
                          >
                            {user.name}
                          </Link>
                        </td>
                        <td className="py-2 px-4">
                          <Link
                            to={userUrl}
                            className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline"
                          >
                            {user.email}
                          </Link>
                        </td>
                        <td className="py-2 px-4">
                          <button
                            onClick={() => deleteUser(user.id, user.name)}
                            className="bg-red-600 hover:bg-red-700 text-white font-medium py-1 px-3 rounded text-sm transition-colors"
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {/* Backend Status */}
        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
            🔧 Backend Status
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 bg-gray-100 dark:bg-gray-700 rounded-lg">
              <h3 className="font-medium text-gray-900 dark:text-white mb-2">
                API Endpoints
              </h3>
              <ul className="text-sm text-gray-600 dark:text-gray-300 space-y-1">
                <li>GET /api/health - Health check</li>
                <li>GET /api/users - List all users</li>
                <li>POST /api/users - Create user</li>
                <li>GET /api/users/:id - Get user by ID</li>
                <li>PUT /api/users/:id - Update user</li>
                <li>DELETE /api/users/:id - Delete user</li>
              </ul>
            </div>
            <div className="p-4 bg-gray-100 dark:bg-gray-700 rounded-lg">
              <h3 className="font-medium text-gray-900 dark:text-white mb-2">
                Server Details
              </h3>
              <ul className="text-sm text-gray-600 dark:text-gray-300 space-y-1">
                <li>Base URL: {API_BASE_URL}</li>
                <li>Database: PostgreSQL</li>
                <li>Framework: Go + Chi Router</li>
                <li>CORS: Enabled for React dev server</li>
              </ul>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
