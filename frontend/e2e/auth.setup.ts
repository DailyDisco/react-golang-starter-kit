import { test as setup } from "@playwright/test";

const authFile = "e2e/.auth/user.json";

/**
 * Authentication setup for E2E tests.
 *
 * This file runs before authenticated tests to:
 * 1. Log in with test credentials
 * 2. Save authentication state to a file
 * 3. Subsequent tests can reuse this state
 *
 * Set these environment variables:
 * - E2E_TEST_USER: Test user email
 * - E2E_TEST_PASSWORD: Test user password
 */
setup("authenticate", async ({ page }) => {
  // Skip if no test credentials are configured
  if (!process.env.E2E_TEST_USER || !process.env.E2E_TEST_PASSWORD) {
    console.log("Skipping authentication setup: E2E_TEST_USER and E2E_TEST_PASSWORD not set");
    console.log("Authenticated tests will be skipped.");
    return;
  }

  // Navigate to login page
  await page.goto("/login");

  // Fill in credentials
  await page.getByLabel("Email").fill(process.env.E2E_TEST_USER);
  await page.getByLabel("Password").fill(process.env.E2E_TEST_PASSWORD);

  // Submit the form
  await page.getByRole("button", { name: /sign in/i }).click();

  // Wait for successful login - should redirect away from login page
  await page.waitForURL((url) => !url.pathname.includes("/login"), {
    timeout: 30000,
  });

  // Save authentication state
  await page.context().storageState({ path: authFile });

  console.log("Authentication setup complete");
});
