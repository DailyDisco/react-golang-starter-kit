import { createFileRoute } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";

import { Button } from "../../components/ui/button";
import { UserService } from "../../services/users/userService";

export const Route = createFileRoute("/(dashboard)/users/$userId")({
  component: UserDetailPage,
  // Validate the userId parameter
  validateSearch: (search) => ({
    tab: search.tab as "profile" | "settings" | undefined,
  }),
  // Loader with parameter validation using real API
  loader: async ({ params, context }) => {
    const userId = Number(params.userId);
    if (isNaN(userId)) {
      throw new Error("Invalid user ID");
    }

    // Fetch user from API using queryClient for caching
    const user = await context.queryClient.ensureQueryData({
      queryKey: ["user", userId],
      queryFn: () => UserService.getUserById(userId),
      staleTime: 60 * 1000, // 1 minute
    });

    return { user };
  },
});

function UserDetailPage() {
  const { userId } = Route.useParams();
  const data = Route.useLoaderData();
  const navigate = Route.useNavigate();

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <div className="mb-6">
        <Button
          variant="outline"
          onClick={() => navigate({ to: "/users" })}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Users
        </Button>
      </div>

      <div className="bg-card rounded-lg border p-6">
        <h1 className="mb-4 text-2xl font-bold">User Details</h1>

        <div className="space-y-3">
          <div>
            <strong>ID:</strong> {data.user.id}
          </div>
          <div>
            <strong>Name:</strong> {data.user.name}
          </div>
          <div>
            <strong>Email:</strong> {data.user.email}
          </div>
          <div>
            <strong>Role:</strong> {data.user.role}
          </div>
        </div>

        <div className="mt-6">
          <Button>Edit User</Button>
        </div>
      </div>
    </div>
  );
}
