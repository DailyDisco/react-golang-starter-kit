import { clearLocalStorage, expect, test, testUsers } from "./fixtures/test-fixtures";

test.describe("Signup Flow - Complete User Journey", () => {
  test.beforeEach(async ({ page }) => {
    await clearLocalStorage(page);
  });

  test.describe("Successful Registration", () => {
    test("should complete full registration with valid data", async ({ page, authPage }) => {
      const uniqueEmail = `e2e-signup-${Date.now()}@example.com`;

      await authPage.register("E2E Test User", uniqueEmail, "ValidPass123!");

      // Should redirect to dashboard or home after successful registration.
      // toHaveURL polls the URL; waitForURL additionally waits for a load
      // event, which the SPA's pushState navigation may not emit in time.
      await expect(page).toHaveURL(/dashboard|\/$/i, { timeout: 15000 });

      // Should see logged-in state: the navbar swaps its Sign in / Sign up links
      // for the user menu once the session is live. (Matching "welcome" or
      // "dashboard" text hit both the transient toast and landing-page copy,
      // which made the assertion strict-mode flaky.)
      await expect(page.getByRole("link", { name: /sign in/i })).toHaveCount(0, { timeout: 10000 });
    });

    test("should show password strength indicator", async ({ page, authPage }) => {
      await authPage.goto("/register");

      const passwordInput = page.getByLabel("Password", { exact: true });

      // Type a weak password
      await passwordInput.fill("weak");
      await expect(page.getByText(/weak/i).or(page.locator('[data-testid="password-strength"]'))).toBeVisible();

      // Type a stronger password
      await passwordInput.fill("StrongPass123!");
      await expect(page.getByText(/strong/i).or(page.locator('[data-testid="password-strength"]'))).toBeVisible();
    });

    test("should auto-focus the first input field", async ({ page, authPage }) => {
      await authPage.goto("/register");

      // The first input (Full Name) should be focused
      const nameInput = page.getByLabel("Full Name");
      await expect(nameInput).toBeFocused({ timeout: 5000 });
    });
  });

  test.describe("Form Validation", () => {
    test("should validate email format", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("Test User");
      await page.getByLabel("Email").fill("not-an-email");
      // The form validates on submit (react-hook-form default mode), so submit
      // rather than blur to surface the error.
      await page.getByRole("button", { name: /create account/i }).click();

      await expect(page.locator("#email-error")).toBeVisible();
      await expect(page.locator("#email-error")).toContainText(/valid email/i);
    });

    test("should show password requirements", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("Test User");
      await page.getByLabel("Email").fill("test@example.com");
      await page.getByLabel("Password", { exact: true }).fill("weak");
      await page.getByLabel("Confirm Password").fill("weak");
      await page.getByRole("button", { name: /create account/i }).click();

      // Should show the password requirements checklist (scoped to the strength
      // meter list item; the inline validation error also repeats the same text)
      await expect(page.getByRole("listitem").filter({ hasText: /at least 8 characters/i })).toBeVisible();
    });

    test("should validate name minimum length", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("A");
      await page.getByLabel("Email").fill("test@example.com");
      await page.getByLabel("Password", { exact: true }).fill("ValidPass123!");
      await page.getByLabel("Confirm Password").fill("ValidPass123!");
      await page.getByRole("button", { name: /create account/i }).click();

      await expect(page.getByText(/at least 2 characters/i)).toBeVisible();
    });

    test("should show error for existing email", async ({ page, authPage }) => {
      // Mock the backend conflict response so the test does not depend on
      // whether this account happens to exist in the ephemeral test database.
      await page.route("**/api/v1/auth/register", (route) => {
        route.fulfill({
          status: 409,
          contentType: "application/json",
          body: JSON.stringify({ error: "CONFLICT", message: "User with this email already exists" }),
        });
      });

      await authPage.register(testUsers.existing.name, testUsers.existing.email, "ValidPass123!");

      // Should show the inline error alert about the existing user
      await expect(page.locator('[data-slot="alert"]')).toBeVisible({ timeout: 10000 });
      await expect(page.locator('[data-slot="alert"]')).toContainText(/already exists/i);
    });
  });

  test.describe("OAuth Registration", () => {
    test("should display OAuth registration options", async ({ page, authPage }) => {
      await authPage.goto("/register");

      // Check for OAuth buttons (Google, GitHub)
      await expect(page.getByRole("button", { name: /sign up with google/i })).toBeVisible();
      await expect(page.getByRole("button", { name: /sign up with github/i })).toBeVisible();
    });
  });

  test.describe("Accessibility", () => {
    test("should have proper form labels", async ({ page, authPage }) => {
      await authPage.goto("/register");

      // All inputs should have associated labels
      await expect(page.getByLabel("Full Name")).toBeVisible();
      await expect(page.getByLabel("Email")).toBeVisible();
      await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
      await expect(page.getByLabel("Confirm Password")).toBeVisible();
    });

    test("should support keyboard navigation", async ({ page, authPage }) => {
      await authPage.goto("/register");

      // Anchor focus on the first field (its autoFocus mount races the first
      // Tab keystroke), then Tab through the form fields.
      await page.getByLabel("Full Name").click();

      await page.keyboard.press("Tab");
      await expect(page.getByLabel("Email")).toBeFocused();

      await page.keyboard.press("Tab");
      await expect(page.getByLabel("Password", { exact: true })).toBeFocused();
    });
  });
});
