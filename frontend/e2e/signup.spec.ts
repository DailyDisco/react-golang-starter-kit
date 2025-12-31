import { clearLocalStorage, expect, test, testUsers } from "./fixtures/test-fixtures";

test.describe("Signup Flow - Complete User Journey", () => {
  test.beforeEach(async ({ page }) => {
    await clearLocalStorage(page);
  });

  test.describe("Successful Registration", () => {
    test("should complete full registration with valid data", async ({ page, authPage }) => {
      const uniqueEmail = `e2e-signup-${Date.now()}@example.com`;

      await authPage.register("E2E Test User", uniqueEmail, "ValidPass123!");

      // Should redirect to dashboard or home after successful registration
      await page.waitForURL(/dashboard|\/$/i, { timeout: 15000 });

      // Should see welcome or logged-in state
      await expect(
        page.getByText(/welcome/i).or(page.getByText(/dashboard/i).or(page.getByRole("button", { name: /user menu/i })))
      ).toBeVisible({ timeout: 10000 });
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
    test("should validate email format in real-time", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("Test User");
      await page.getByLabel("Email").fill("not-an-email");
      await page.getByLabel("Email").blur();

      await expect(page.getByText(/valid email/i)).toBeVisible();
    });

    test("should show password requirements", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await page.getByLabel("Full Name").fill("Test User");
      await page.getByLabel("Email").fill("test@example.com");
      await page.getByLabel("Password", { exact: true }).fill("weak");
      await page.getByLabel("Confirm Password").fill("weak");
      await page.getByRole("button", { name: /create account/i }).click();

      // Should show password requirements
      await expect(
        page
          .getByText(/at least 8 characters/i)
          .or(page.getByText(/uppercase/i))
          .or(page.getByText(/lowercase/i))
          .or(page.getByText(/number/i))
      ).toBeVisible();
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
      // Try to register with an existing email
      await authPage.register(testUsers.existing.name, testUsers.existing.email, "ValidPass123!");

      // Should show error about existing user
      await expect(
        page
          .getByText(/already exists/i)
          .or(page.getByText(/already registered/i))
          .or(page.getByRole("alert"))
      ).toBeVisible({ timeout: 10000 });
    });
  });

  test.describe("Terms and Privacy", () => {
    test("should display terms of service link", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await expect(page.getByRole("link", { name: /terms/i })).toBeVisible();
    });

    test("should display privacy policy link", async ({ page, authPage }) => {
      await authPage.goto("/register");

      await expect(page.getByRole("link", { name: /privacy/i })).toBeVisible();
    });
  });

  test.describe("OAuth Registration", () => {
    test("should display OAuth registration options", async ({ page, authPage }) => {
      await authPage.goto("/register");

      // Check for OAuth buttons
      await expect(
        page
          .getByRole("button", { name: /google/i })
          .or(page.getByRole("button", { name: /github/i }).or(page.locator("text=Continue with")))
      ).toBeVisible();
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

      // Tab through form fields
      await page.keyboard.press("Tab");
      await expect(page.getByLabel("Full Name")).toBeFocused();

      await page.keyboard.press("Tab");
      await expect(page.getByLabel("Email")).toBeFocused();

      await page.keyboard.press("Tab");
      await expect(page.getByLabel("Password", { exact: true })).toBeFocused();
    });
  });
});
