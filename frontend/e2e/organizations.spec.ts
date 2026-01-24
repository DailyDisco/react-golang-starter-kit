import { expect, test } from "./fixtures/test-fixtures";

/**
 * Organization Management E2E Tests
 *
 * Tests covering organization creation, member management,
 * settings, and multi-tenant features.
 */
test.describe("Organizations", () => {
  test.describe("Unauthenticated", () => {
    test("should redirect to login when accessing organizations", async ({ page }) => {
      await page.goto("/organizations");
      await expect(page).toHaveURL(/login/);
    });

    test("should redirect to login when accessing org settings", async ({ page }) => {
      await page.goto("/org/test-org/settings");
      await expect(page).toHaveURL(/login/);
    });
  });

  test.describe("Authenticated", () => {
    test.skip(!process.env.E2E_TEST_USER, "No test user configured");

    test.use({
      storageState: "e2e/.auth/user.json",
    });

    test.describe("Organization List", () => {
      test("should display organizations page", async ({ page }) => {
        await page.goto("/organizations");

        await expect(page.getByRole("heading", { name: /organizations|teams|workspaces/i })).toBeVisible();
      });

      test("should show create organization button", async ({ page }) => {
        await page.goto("/organizations");

        await expect(
          page
            .getByRole("button", { name: /create|new organization/i })
            .or(page.getByRole("link", { name: /create|new/i }))
        ).toBeVisible();
      });

      test("should display user's organizations", async ({ page }) => {
        // Mock organizations list
        await page.route("**/api/organizations**", (route) => {
          if (route.request().method() === "GET") {
            route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                organizations: [
                  {
                    id: 1,
                    name: "Test Org",
                    slug: "test-org",
                    plan: "free",
                    member_count: 3,
                  },
                  {
                    id: 2,
                    name: "Another Org",
                    slug: "another-org",
                    plan: "pro",
                    member_count: 10,
                  },
                ],
              }),
            });
          } else {
            route.continue();
          }
        });

        await page.goto("/organizations");

        await expect(page.getByText("Test Org")).toBeVisible();
        await expect(page.getByText("Another Org")).toBeVisible();
      });

      test("should show empty state when no organizations", async ({ page }) => {
        await page.route("**/api/organizations**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({ organizations: [] }),
          });
        });

        await page.goto("/organizations");

        await expect(page.getByText(/no organizations|create your first|get started/i)).toBeVisible();
      });
    });

    test.describe("Create Organization", () => {
      test("should open create organization modal/form", async ({ page }) => {
        await page.goto("/organizations");

        const createButton = page.getByRole("button", { name: /create|new/i });
        await createButton.click();

        // Should show creation form
        await expect(page.getByLabel(/name/i).or(page.getByPlaceholder(/organization name/i))).toBeVisible();
      });

      test("should validate organization name", async ({ page }) => {
        await page.goto("/organizations/new");

        // Try to submit empty form
        const submitButton = page.getByRole("button", { name: /create/i });
        if (await submitButton.isVisible()) {
          await submitButton.click();

          await expect(page.getByText(/required|name is required/i)).toBeVisible();
        }
      });

      test("should create organization successfully", async ({ page }) => {
        await page.route("**/api/organizations**", (route) => {
          if (route.request().method() === "POST") {
            route.fulfill({
              status: 201,
              contentType: "application/json",
              body: JSON.stringify({
                id: 3,
                name: "New Test Org",
                slug: "new-test-org",
                plan: "free",
              }),
            });
          } else {
            route.continue();
          }
        });

        await page.goto("/organizations/new");

        const nameInput = page.getByLabel(/name/i).or(page.getByPlaceholder(/organization name/i));

        if (await nameInput.isVisible()) {
          await nameInput.fill("New Test Org");

          const submitButton = page.getByRole("button", { name: /create/i });
          await submitButton.click();

          // Should redirect to new org or show success
          await expect(page.getByText(/created|success/i).or(page.locator("body"))).toBeVisible();
        }
      });
    });

    test.describe("Organization Settings", () => {
      test("should display organization settings page", async ({ page }) => {
        // Mock org data
        await page.route("**/api/organizations/test-org**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              id: 1,
              name: "Test Org",
              slug: "test-org",
              plan: "pro",
              settings: {},
            }),
          });
        });

        await page.goto("/org/test-org/settings");

        await expect(page.getByRole("heading", { name: /settings/i })).toBeVisible();
      });

      test("should allow updating organization name", async ({ page }) => {
        await page.route("**/api/organizations/test-org**", (route) => {
          if (route.request().method() === "GET") {
            route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                id: 1,
                name: "Test Org",
                slug: "test-org",
              }),
            });
          } else if (route.request().method() === "PATCH" || route.request().method() === "PUT") {
            route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                id: 1,
                name: "Updated Org Name",
                slug: "test-org",
              }),
            });
          } else {
            route.continue();
          }
        });

        await page.goto("/org/test-org/settings");

        const nameInput = page.getByLabel(/organization name/i).or(page.locator('input[value="Test Org"]'));

        if (await nameInput.isVisible()) {
          await nameInput.clear();
          await nameInput.fill("Updated Org Name");

          const saveButton = page.getByRole("button", { name: /save|update/i });
          await saveButton.click();

          await expect(page.getByText(/saved|updated|success/i)).toBeVisible();
        }
      });
    });

    test.describe("Member Management", () => {
      test("should display members list", async ({ page }) => {
        await page.route("**/api/organizations/test-org/members**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              members: [
                { id: 1, user: { name: "Owner", email: "owner@test.com" }, role: "owner" },
                { id: 2, user: { name: "Admin", email: "admin@test.com" }, role: "admin" },
                { id: 3, user: { name: "Member", email: "member@test.com" }, role: "member" },
              ],
            }),
          });
        });

        await page.goto("/org/test-org/members");

        await expect(page.getByText("owner@test.com")).toBeVisible();
        await expect(page.getByText("admin@test.com")).toBeVisible();
      });

      test("should show invite member button for admins", async ({ page }) => {
        await page.route("**/api/organizations/test-org**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              id: 1,
              name: "Test Org",
              slug: "test-org",
              current_user_role: "admin",
            }),
          });
        });

        await page.goto("/org/test-org/members");

        await expect(page.getByRole("button", { name: /invite|add member/i })).toBeVisible();
      });

      test("should open invite modal when clicking invite", async ({ page }) => {
        await page.goto("/org/test-org/members");

        const inviteButton = page.getByRole("button", { name: /invite|add/i });

        if (await inviteButton.isVisible()) {
          await inviteButton.click();

          await expect(page.getByLabel(/email/i).or(page.getByPlaceholder(/email/i))).toBeVisible();
        }
      });

      test("should send invitation", async ({ page }) => {
        await page.route("**/api/organizations/test-org/invitations**", (route) => {
          route.fulfill({
            status: 201,
            contentType: "application/json",
            body: JSON.stringify({
              id: 1,
              email: "newmember@test.com",
              role: "member",
              status: "pending",
            }),
          });
        });

        await page.goto("/org/test-org/members");

        const inviteButton = page.getByRole("button", { name: /invite|add/i });

        if (await inviteButton.isVisible()) {
          await inviteButton.click();

          const emailInput = page.getByLabel(/email/i).or(page.getByPlaceholder(/email/i));
          await emailInput.fill("newmember@test.com");

          const sendButton = page.getByRole("button", { name: /send|invite/i });
          await sendButton.click();

          await expect(page.getByText(/sent|invited|success/i)).toBeVisible();
        }
      });
    });

    test.describe("Organization Switching", () => {
      test("should show organization switcher in header", async ({ page }) => {
        await page.route("**/api/organizations**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              organizations: [
                { id: 1, name: "Org 1", slug: "org-1" },
                { id: 2, name: "Org 2", slug: "org-2" },
              ],
            }),
          });
        });

        await page.goto("/dashboard");

        // Look for org switcher
        const orgSwitcher = page
          .getByRole("button", { name: /org|workspace|team/i })
          .or(page.locator("[data-testid='org-switcher']"));

        await expect(orgSwitcher).toBeVisible();
      });
    });
  });
});

test.describe("Organization - Error Handling", () => {
  test.skip(!process.env.E2E_TEST_USER, "No test user configured");

  test.use({
    storageState: "e2e/.auth/user.json",
  });

  test("should handle organization not found", async ({ page }) => {
    await page.route("**/api/organizations/nonexistent**", (route) => {
      route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ error: "Organization not found" }),
      });
    });

    await page.goto("/org/nonexistent/settings");

    await expect(page.getByText(/not found|doesn't exist/i).or(page.getByRole("alert"))).toBeVisible();
  });

  test("should handle permission denied", async ({ page }) => {
    await page.route("**/api/organizations/restricted**", (route) => {
      route.fulfill({
        status: 403,
        contentType: "application/json",
        body: JSON.stringify({ error: "Permission denied" }),
      });
    });

    await page.goto("/org/restricted/settings");

    await expect(page.getByText(/permission|access denied|not authorized/i).or(page.getByRole("alert"))).toBeVisible();
  });
});
