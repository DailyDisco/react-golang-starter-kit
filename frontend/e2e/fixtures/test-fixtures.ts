import { test as base, expect, type Page } from "@playwright/test";

// Test user data
export const testUsers = {
  existing: {
    email: "test@example.com",
    password: "TestPassword123!",
    name: "Test User",
  },
  new: {
    email: `e2e-${Date.now()}@example.com`,
    password: "NewPassword123!",
    name: "New E2E User",
  },
};

// Page objects for common actions
export class AuthPage {
  constructor(private page: Page) {}

  async goto(path: "/login" | "/register") {
    await this.page.goto(path);
  }

  async login(email: string, password: string) {
    await this.page.goto("/login");
    await this.page.getByLabel("Email").fill(email);
    await this.page.getByLabel("Password").fill(password);
    await this.page.getByRole("button", { name: /sign in/i }).click();
  }

  async register(name: string, email: string, password: string) {
    await this.page.goto("/register");
    await this.page.getByLabel("Full Name").fill(name);
    await this.page.getByLabel("Email").fill(email);
    await this.page.getByLabel("Password", { exact: true }).fill(password);
    await this.page.getByLabel("Confirm Password").fill(password);
    await this.page.getByRole("button", { name: /create account/i }).click();
  }

  async logout() {
    // Look for user menu button and click logout
    const userMenu = this.page.getByRole("button", { name: /user menu/i });
    if (await userMenu.isVisible()) {
      await userMenu.click();
      await this.page.getByRole("menuitem", { name: /logout|sign out/i }).click();
    }
  }

  async expectLoggedIn() {
    // User should see protected content or user menu
    await expect(
      this.page.getByRole("button", { name: /user menu/i }).or(this.page.getByText(/dashboard/i))
    ).toBeVisible({ timeout: 10000 });
  }

  async expectLoggedOut() {
    // Should see login button or be on login page
    await expect(
      this.page.getByRole("link", { name: /sign in|login/i }).or(this.page.getByRole("heading", { name: /sign in/i }))
    ).toBeVisible({ timeout: 10000 });
  }

  async expectError(message?: string | RegExp) {
    const alert = this.page.getByRole("alert");
    await expect(alert).toBeVisible();
    if (message) {
      await expect(alert).toContainText(message);
    }
  }

  async expectValidationError(fieldName: string, message?: string | RegExp) {
    const field = this.page.getByLabel(fieldName);
    const errorText = this.page.locator(`text=${message}`);

    if (message) {
      await expect(errorText.or(field.locator("~ p.text-red-500"))).toBeVisible();
    }
  }
}

// Custom test fixture with page objects
type Fixtures = {
  authPage: AuthPage;
};

export const test = base.extend<Fixtures>({
  authPage: async ({ page }, use) => {
    const authPage = new AuthPage(page);
    await use(authPage);
  },
});

export { expect } from "@playwright/test";

// Helper functions
export async function waitForApiResponse(page: Page, urlPattern: string | RegExp): Promise<void> {
  await page.waitForResponse((response) => {
    const url = response.url();
    if (typeof urlPattern === "string") {
      return url.includes(urlPattern);
    }
    return urlPattern.test(url);
  });
}

export async function clearLocalStorage(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
}

export async function setAuthToken(page: Page, token: string): Promise<void> {
  await page.evaluate((t) => {
    localStorage.setItem("auth_token", t);
  }, token);
}
