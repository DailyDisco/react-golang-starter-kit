import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export function AuthButtons() {
  return (
    <div className="flex items-center space-x-2">
      <Button
        variant="ghost"
        asChild
      >
        <Link
          to="/login"
          search={{}}
        >
          Sign in
        </Link>
      </Button>
      <Button asChild>
        <Link
          to="/register"
          search={{}}
        >
          Sign up
        </Link>
      </Button>
    </div>
  );
}
