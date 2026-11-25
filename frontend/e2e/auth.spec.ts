import { clearLocalStorage, expect, test, testUsers } from "./fixtures/test-fixtures";

test.describe("Authentication", () => {
  test.beforeEach(async ({ page }) => {
    // Clear any existing auth state before each test
    await clearLocalStorage(page);
  });

  test.describe("Login Flow", () => {
    test("should display login form", async ({ page, authPage }) => {
      await authPage.goto("/login");

      await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
      await expect(page.getByLabel("Email")).toBeVisible();
      // exact: the "Show password" toggle button also matches the substring query
      await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
      await expect(page.getByRole("button", { name: /sign in/i })).toBeVisible();
    });

    test("should show validation errors for empty fields", async ({ page, authPage }) => {
      await authPage.goto("/login");
      await page.getByRole("button", { name: /sign in/i }).click();

      // Check for validation error messages on both fields
      await expect(page.locator("#email-error")).toContainText(/valid email/i);
      await expect(page.locator("#password-error")).toContainText(/required/i);
    });

    test("should show validation error for invalid email", async ({ page, authPage }) => {
      await authPage.goto("/login");
      await page.getByLabel("Email").fill("invalid-email");
      // exact: the "Show password" toggle button also matches the substring query
      await page.getByLabel("Password", { exact: true }).fill("password123");
      await page.getByRole("button", { name: /sign in/i }).click();

      await expect(page.getByText(/valid email/i)).toBeVisible();
    });

    test("should show error for incorrect credentials", async ({ page, authPage }) => {
      await authPage.goto("/login");
      await page.getByLabel("Email").fill("wrong@example.com");
      // exact: the "Show password" toggle button also matches the substring query
      await page.getByLabel("Password", { exact: true }).fill("wrongpassword");
      await page.getByRole("button", { name: /sign in/i }).click();

      // Wait for API response and check for error
      await authPage.expectError();
    });

    test("should toggle password visibility", async ({ page, authPage }) => {
      await authPage.goto("/login");

      // exact: the "Show password" toggle button also matches the substring query
      const passwordInput = page.getByLabel("Password", { exact: true });
      const toggleButton = page.getByRole("button", { name: /show password|hide password/i });

      // Initially password should be hidden
      await expect(passwordInput).toHaveAttribute("type", "password");

      // Click toggle to show password
      await toggleButton.first().click();
      await expect(passwordInput).toHaveAttribute("type", "text");

      // Click toggle to hide password again
      await toggleButton.first().click();
      await expect(passwordInput).toHaveAttribute("type", "password");
    });

    test("should navigate to register page", async ({ page, authPage }) => {
      await authPage.goto("/login");
      // Scope to the form footer link: the site navbar also has a "Sign up" link
      await page
        .getByRole("main")
        .getByRole("link", { name: /sign up/i })
        .click();

      await expect(page).toHaveURL(/register/);
      await expect(page.getByRole("heading", { name: /create account/i })).toBeVisible();
    });

    test("should disable form during submission", async ({ page, authPage }) => {
      await authPage.goto("/login");

      // Hold the login request open so the in-flight (disabled) button state
      // is observable deterministically instead of racing the response.
      await page.route("**/api/v1/auth/login", async (route) => {
        await new Promise<void>((resolve) => setTimeout(resolve, 1500));
        await route.continue();
      });

      await page.getByLabel("Email").fill("test@example.com");
      // exact: the "Show password" toggle button also matches the substring query
      await page.getByLabel("Password", { exact: true }).fill("password123");

      // Click submit and immediately check that form is disabled
      const submitButton = page.getByRole("button", { name: /sign in/i });
      await submitButton.click();

      // Button should show loading state
      await expect(submitButton).toBeDisabled();
    });
  });

  test.describe("Registration Flow", () => {
    test("should display registration form", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await expect(page.getByRole("heading", { name: /create account/i })).toBeVisible();
      await expect(page.getByLabel("Full Name")).toBeVisible();
      await expect(page.getByLabel("Email")).toBeVisible();
      await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
      await expect(page.getByLabel("Confirm Password")).toBeVisible();
      await expect(page.getByRole("button", { name: /create account/i })).toBeVisible();
    });

    test("should show validation errors for empty fields", async ({ page, authPage }) => {
      await authPage.goto("/register");
      await page.getByRole("button", { name: /create account/i }).click();

      // Check for validation error messages
      await expect(page.getByText(/at least 2 characters/i).or(page.getByText(/required/i))).toBeVisible();
    });

    test("should show error for short password", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("Test User");
      await page.getByLabel("Email").fill("test@example.com");
      await page.getByLabel("Password", { exact: true }).fill("short");
      await page.getByLabel("Confirm Password").fill("short");
      await page.getByRole("button", { name: /create account/i }).click();

      // The inline error for the password field (the strength meter shows the
      // same requirement text as a checklist item, hence the scoping)
      await expect(page.locator("#password-error")).toContainText(/at least 8 characters/i);
    });

    test("should show error for mismatched passwords", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("Test User");
      await page.getByLabel("Email").fill("test@example.com");
      await page.getByLabel("Password", { exact: true }).fill("Password123!");
      await page.getByLabel("Confirm Password").fill("Different123!");
      await page.getByRole("button", { name: /create account/i }).click();

      await expect(page.getByText(/passwords don't match/i)).toBeVisible();
    });

    test("should navigate to login page", async ({ page, authPage }) => {
      await authPage.goto("/register");
      // Scope to the form footer link: the site navbar also has a "Sign in" link
      await page
        .getByRole("main")
        .getByRole("link", { name: /sign in/i })
        .click();

      await expect(page).toHaveURL(/login/);
      await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
    });

    test("should toggle password visibility", async ({ page, authPage }) => {
      await authPage.goto("/register");

      const passwordInput = page.getByLabel("Password", { exact: true });

      // Initially password should be hidden
      await expect(passwordInput).toHaveAttribute("type", "password");

      // Find the password visibility toggle (aria-labelled) and click it
      const toggleButtons = page.getByRole("button", { name: /show password|hide password/i });
      await toggleButtons.first().click();
      await expect(passwordInput).toHaveAttribute("type", "text");
    });
  });

  test.describe("Protected Routes", () => {
    test("should redirect to login when accessing protected route while unauthenticated", async ({ page }) => {
      await page.goto("/dashboard");

      // Should redirect to login
      await expect(page).toHaveURL(/login/);
    });

    test("should redirect to login when accessing user settings while unauthenticated", async ({ page }) => {
      await page.goto("/settings");

      // Should redirect to login
      await expect(page).toHaveURL(/login/);
    });
  });

  test.describe("Session Persistence", () => {
    test("should maintain session after page refresh", async ({ page }) => {
      // This test requires a valid logged-in session
      // Skip if we don't have test credentials set up
      test.skip(!process.env.E2E_TEST_USER, "No test user configured");

      // Login
      await page.goto("/login");
      await page.getByLabel("Email").fill(process.env.E2E_TEST_USER || testUsers.existing.email);
      // exact: the "Show password" toggle button also matches the substring query
      await page
        .getByLabel("Password", { exact: true })
        .fill(process.env.E2E_TEST_PASSWORD || testUsers.existing.password);
      await page.getByRole("button", { name: /sign in/i }).click();

      // Wait for successful login
      await page.waitForURL(/dashboard|\/$/);

      // Refresh page
      await page.reload();

      // Should still be logged in
      await expect(page).not.toHaveURL(/login/);
    });
  });
});

test.describe("OAuth Buttons", () => {
  test("should display OAuth login buttons", async ({ page, authPage }) => {
    await authPage.goto("/login");

    // Check for OAuth buttons (Google, GitHub)
    await expect(page.getByRole("button", { name: /continue with google/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /continue with github/i })).toBeVisible();
  });
});
