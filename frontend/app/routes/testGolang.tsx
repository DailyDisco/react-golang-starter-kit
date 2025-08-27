import { PencilIcon, TrashIcon } from "lucide-react";
import React, { useState } from "react";
import { useUsers } from "../hooks/use-users";
import { UserForm } from "../components/UserForm";
import { Button } from "../components/ui/button";

interface User {
  ID: number;
  name: string;
  email: string;
}

const TestGolang = () => {
  const { users, loading, error, addUser, editUser, removeUser } = useUsers();
  const [editingUser, setEditingUser] = useState<User | null>(null);

  const handleAddOrUpdateUser = async (
    name: string,
    email: string,
    id?: number,
  ) => {
    if (id) {
      await editUser({ ID: id, name, email });
    } else {
      await addUser(name, email);
    }
    setEditingUser(null); // Clear editing state after submission
  };

  const handleDelete = async (id: number) => {
    await removeUser(id);
  };

  return (
    <div className="max-w-md mx-auto mt-12 p-6 bg-white rounded-xl shadow border">
      <h1 className="text-2xl font-bold text-center mb-6">Test Golang</h1>

      {loading && <p className="text-center text-blue-500">Loading users...</p>}
      {error && <p className="text-center text-red-500">Error: {error}</p>}

      <div className="mb-6">
        {users.length > 0 ? (
          <ul className="space-y-2">
            {users.map((user) => (
              <li
                key={user.ID}
                className="flex items-center justify-between border rounded-lg p-3 bg-gray-50 hover:bg-gray-100 transition"
              >
                <div>
                  <span className="block font-medium text-gray-900">
                    {user.name}
                  </span>
                  <span className="block text-sm text-gray-500">
                    {user.email}
                  </span>
                </div>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setEditingUser(user)}
                    aria-label={`Edit user ${user.name}`}
                  >
                    <PencilIcon className="w-5 h-5 text-yellow-500" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDelete(user.ID)}
                    aria-label={`Delete user ${user.name}`}
                  >
                    <TrashIcon className="w-5 h-5 text-red-500" />
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        ) : (
          !loading && (
            <p className="text-center text-gray-500">No users found</p>
          )
        )}
      </div>

      <h2 className="text-xl font-bold mb-4">
        {editingUser ? "Edit User" : "Add New User"}
      </h2>
      <UserForm
        onSubmit={handleAddOrUpdateUser}
        initialData={editingUser}
        isLoading={loading}
      />
    </div>
  );
};

export default TestGolang;
