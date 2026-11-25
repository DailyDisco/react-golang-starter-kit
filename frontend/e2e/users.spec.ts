import { expect, test } from "./fixtures/test-fixtures";

/**
 * User Management E2E Tests
 *
 * These tests cover the user listing and detail pages.
 * Note: Some tests require authentication and will be skipped
 * if E2E_TEST_USER/E2E_TEST_PASSWORD are not set.
 */
test.describe("User Management", () => {
  test.describe("Users List Page", () => {
    test("should redirect unauthenticated users to login", async ({ page }) => {
      // The app's users page lives at /admin/users (the app route group is
      // auth-guarded); /users itself is not a route.
      await page.goto("/admin/users");

      // Should redirect to login
      await expect(page).toHaveURL(/login/);
    });

    test.describe("Authenticated", () => {
      test.skip(!process.env.E2E_TEST_USER, "No test user configured");

      test.use({
        storageState: "e2e/.auth/user.json",
      });

      test("should display users page with header", async ({ page }) => {
        await page.goto("/admin/users");

        await expect(page.getByRole("heading", { name: /user management/i })).toBeVisible();
        await expect(page.getByText(/manage users, roles, and permissions/i)).toBeVisible();
      });

      test("should show loading state while fetching users", async ({ page }) => {
        // Slow down the response to see the in-flight loading indicator
        // (setTimeout rather than Promise.withResolvers: tsconfig lib is ES2022)
        await page.route("**/api/v1/users**", async (route) => {
          await new Promise<void>((resolve) => setTimeout(resolve, 1000));
          await route.continue();
        });

        await page.goto("/admin/users");

        // The page shows a spinner while the query is pending
        await expect(page.locator('[class*="animate-spin"]')).toBeVisible();
      });

      test("should display user list when users exist", async ({ page }) => {
        await page.route("**/api/v1/users**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              success: true,
              data: {
                users: [
                  {
                    id: 2,
                    name: "Mocked User",
                    email: "mocked@example.com",
                    role: "user",
                    is_active: true,
                    email_verified: false,
                    created_at: "2024-01-01T00:00:00Z",
                  },
                ],
                count: 1,
                total: 1,
                page: 1,
                limit: 20,
                total_pages: 1,
              },
            }),
          });
        });

        await page.goto("/admin/users");

        // The mocked user should appear in the table
        await expect(page.getByText("Mocked User")).toBeVisible({ timeout: 10000 });
        await expect(page.getByText("mocked@example.com")).toBeVisible();
      });

      test("should show empty state when no users exist", async ({ page }) => {
        // Mock empty user list
        await page.route("**/api/v1/users**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              success: true,
              data: { users: [], count: 0, total: 0, page: 1, limit: 20, total_pages: 0 },
            }),
          });
        });

        await page.goto("/admin/users");

        await expect(page.getByText(/no users found/i)).toBeVisible({ timeout: 10000 });
      });
    });
  });

  // NOTE: The app has no user-detail route ("/users/:id" is not implemented —
  // the admin users table only offers "User details page coming soon"), so the
  // former "User Detail Page" suite was removed along with the feature it tested.
});

test.describe("Demo Page - User CRUD", () => {
  test("should display demo page", async ({ page }) => {
    await page.goto("/demo");

    // The demo page's heading is its actual title, not the word "demo"
    await expect(page.getByRole("heading", { name: /interactive feature tour/i })).toBeVisible();
  });

  // NOTE: The demo page has no user-creation form (its steps cover feature
  // highlights and file storage), so the former authenticated "User CRUD"
  // tests were removed with the feature they asserted.
});
