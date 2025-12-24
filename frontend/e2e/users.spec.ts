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
      await page.goto("/users");

      // Should redirect to login
      await expect(page).toHaveURL(/login/);
    });

    test.describe("Authenticated", () => {
      test.skip(!process.env.E2E_TEST_USER, "No test user configured");

      test.use({
        storageState: "e2e/.auth/user.json",
      });

      test("should display users page with header", async ({ page }) => {
        await page.goto("/users");

        await expect(page.getByRole("heading", { name: /users/i })).toBeVisible();
        await expect(page.getByText(/manage user accounts/i)).toBeVisible();
      });

      test("should show loading skeleton while fetching users", async ({ page }) => {
        // Slow down the response to see the skeleton
        await page.route("**/api/users**", async (route) => {
          await new Promise((resolve) => setTimeout(resolve, 1000));
          await route.continue();
        });

        await page.goto("/users");

        // Should show skeleton loader
        await expect(page.locator('[class*="skeleton"]').or(page.locator('[class*="animate-pulse"]'))).toBeVisible();
      });

      test("should display user list when users exist", async ({ page }) => {
        await page.goto("/users");

        // Wait for loading to complete
        await page.waitForLoadState("networkidle");

        // Should show user cards or empty state
        const userCards = page.locator("article, [data-testid='user-card']").or(page.getByRole("article"));
        const emptyState = page.getByText(/no users found/i);

        await expect(userCards.first().or(emptyState)).toBeVisible({ timeout: 10000 });
      });

      test("should show empty state when no users exist", async ({ page }) => {
        // Mock empty user list
        await page.route("**/api/users**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({ success: true, data: [] }),
          });
        });

        await page.goto("/users");

        await expect(page.getByText(/no users found/i)).toBeVisible();
        await expect(page.getByRole("link", { name: /create user/i })).toBeVisible();
      });

      test("should navigate to user detail when clicking View Details", async ({ page }) => {
        await page.goto("/users");

        // Wait for users to load
        await page.waitForLoadState("networkidle");

        // Click on first "View Details" button if available
        const viewDetailsButton = page.getByRole("link", { name: /view details/i }).first();

        // Check if there are users to view
        if (await viewDetailsButton.isVisible()) {
          await viewDetailsButton.click();
          await expect(page).toHaveURL(/users\/\d+/);
        }
      });
    });
  });

  test.describe("User Detail Page", () => {
    test("should redirect unauthenticated users to login", async ({ page }) => {
      await page.goto("/users/1");

      // Should redirect to login
      await expect(page).toHaveURL(/login/);
    });

    test.describe("Authenticated", () => {
      test.skip(!process.env.E2E_TEST_USER, "No test user configured");

      test.use({
        storageState: "e2e/.auth/user.json",
      });

      test("should display user details", async ({ page }) => {
        // Mock a specific user
        await page.route("**/api/users/1**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              success: true,
              data: {
                id: 1,
                name: "Test User",
                email: "test@example.com",
                role: "user",
              },
            }),
          });
        });

        await page.goto("/users/1");

        await expect(page.getByRole("heading", { name: /user details/i })).toBeVisible();
        await expect(page.getByText("Test User")).toBeVisible();
        await expect(page.getByText("test@example.com")).toBeVisible();
      });

      test("should show back button that navigates to users list", async ({ page }) => {
        await page.route("**/api/users/1**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              success: true,
              data: {
                id: 1,
                name: "Test User",
                email: "test@example.com",
                role: "user",
              },
            }),
          });
        });

        await page.goto("/users/1");

        const backButton = page.getByRole("button", { name: /back to users/i });
        await expect(backButton).toBeVisible();

        await backButton.click();
        await expect(page).toHaveURL(/users$/);
      });

      test("should show error for invalid user ID", async ({ page }) => {
        await page.goto("/users/invalid");

        // Should show error or redirect
        await expect(page.getByText(/invalid|error|not found/i)).toBeVisible({ timeout: 10000 });
      });

      test("should show not found for non-existent user", async ({ page }) => {
        await page.route("**/api/users/99999**", (route) => {
          route.fulfill({
            status: 404,
            contentType: "application/json",
            body: JSON.stringify({
              success: false,
              error: "User not found",
            }),
          });
        });

        await page.goto("/users/99999");

        // Should show error message
        await expect(page.getByText(/not found|error/i)).toBeVisible({ timeout: 10000 });
      });
    });
  });
});

test.describe("Demo Page - User CRUD", () => {
  test("should display demo page", async ({ page }) => {
    await page.goto("/demo");

    await expect(page.getByRole("heading", { name: /demo/i })).toBeVisible();
  });

  test.describe("Authenticated", () => {
    test.skip(!process.env.E2E_TEST_USER, "No test user configured");

    test.use({
      storageState: "e2e/.auth/user.json",
    });

    test("should show user creation form", async ({ page }) => {
      await page.goto("/demo");

      // Look for user creation form elements
      await expect(page.getByPlaceholder(/name/i).or(page.getByLabel(/name/i))).toBeVisible();
      await expect(page.getByPlaceholder(/email/i).or(page.getByLabel(/email/i))).toBeVisible();
    });

    test("should validate user creation form", async ({ page }) => {
      await page.goto("/demo");

      // Try to submit without filling fields
      const createButton = page.getByRole("button", { name: /create user/i });
      if (await createButton.isVisible()) {
        await createButton.click();

        // Should show validation error
        await expect(page.getByText(/required|fill|validation/i)).toBeVisible();
      }
    });
  });
});
