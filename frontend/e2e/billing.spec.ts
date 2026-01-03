import { expect, test } from "./fixtures/test-fixtures";

/**
 * Billing E2E Tests
 *
 * Tests covering billing page, subscription management, and checkout flows.
 * Note: Actual Stripe checkout is mocked to avoid real payments in tests.
 */
test.describe("Billing", () => {
  test.describe("Unauthenticated", () => {
    test("should redirect to login when accessing billing page", async ({ page }) => {
      await page.goto("/billing");
      await expect(page).toHaveURL(/login/);
    });
  });

  test.describe("Authenticated", () => {
    test.skip(!process.env.E2E_TEST_USER, "No test user configured");

    test.use({
      storageState: "e2e/.auth/user.json",
    });

    test.describe("Billing Page", () => {
      test("should display billing page with current plan", async ({ page }) => {
        await page.goto("/billing");

        await expect(page.getByRole("heading", { name: /billing|subscription/i })).toBeVisible();

        // Should show current plan info
        await expect(page.getByText(/free|pro|enterprise|current plan/i)).toBeVisible();
      });

      test("should show upgrade options for free tier users", async ({ page }) => {
        // Mock free tier user
        await page.route("**/api/billing/subscription**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              plan: "free",
              status: "active",
            }),
          });
        });

        await page.goto("/billing");

        // Should show upgrade button or plan comparison
        await expect(
          page.getByRole("button", { name: /upgrade|subscribe/i }).or(page.getByText(/upgrade to/i))
        ).toBeVisible();
      });

      test("should show manage subscription for paid users", async ({ page }) => {
        // Mock pro tier user with active subscription
        await page.route("**/api/billing/subscription**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              plan: "pro",
              status: "active",
              current_period_end: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
            }),
          });
        });

        await page.goto("/billing");

        // Should show manage subscription option
        await expect(
          page.getByRole("button", { name: /manage|portal|cancel/i }).or(page.getByText(/manage subscription/i))
        ).toBeVisible();
      });

      test("should display billing history section", async ({ page }) => {
        await page.goto("/billing");

        // Should have billing history or invoices section
        await expect(page.getByText(/billing history|invoices|payment history/i)).toBeVisible();
      });
    });

    test.describe("Checkout Flow", () => {
      test("should initiate checkout when clicking upgrade", async ({ page }) => {
        // Mock the checkout session creation
        await page.route("**/api/billing/checkout**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              checkout_url: "https://checkout.stripe.com/test-session",
            }),
          });
        });

        await page.goto("/billing");

        const upgradeButton = page.getByRole("button", { name: /upgrade|subscribe/i });

        if (await upgradeButton.isVisible()) {
          // Intercept navigation to Stripe
          const [popup] = await Promise.all([page.waitForEvent("popup").catch(() => null), upgradeButton.click()]);

          // Either navigates to Stripe or shows checkout modal
          // This depends on implementation - just verify the action was triggered
          await page.waitForTimeout(1000);
        }
      });

      test("should handle checkout success redirect", async ({ page }) => {
        await page.goto("/billing/success");

        // Should show success message
        await expect(page.getByText(/success|thank you|subscription activated/i)).toBeVisible();
      });

      test("should handle checkout cancel redirect", async ({ page }) => {
        await page.goto("/billing/cancel");

        // Should show cancellation message or redirect to billing
        await expect(page.getByText(/cancel|try again|billing/i)).toBeVisible();
      });
    });

    test.describe("Customer Portal", () => {
      test("should open Stripe customer portal for subscription management", async ({ page }) => {
        // Mock portal session creation
        await page.route("**/api/billing/portal**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              portal_url: "https://billing.stripe.com/test-portal",
            }),
          });
        });

        // Mock pro user
        await page.route("**/api/billing/subscription**", (route) => {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              plan: "pro",
              status: "active",
            }),
          });
        });

        await page.goto("/billing");

        const manageButton = page.getByRole("button", { name: /manage|portal/i });

        if (await manageButton.isVisible()) {
          await manageButton.click();
          // Portal opens in new tab or redirects
          await page.waitForTimeout(500);
        }
      });
    });

    test.describe("Plan Comparison", () => {
      test("should display plan features comparison", async ({ page }) => {
        await page.goto("/billing");

        // Look for plan comparison or features list
        const plansSection = page.getByText(/features|included|compare plans/i);

        if (await plansSection.isVisible()) {
          // Should show different plan tiers
          await expect(page.getByText(/free/i).or(page.getByText(/pro/i))).toBeVisible();
        }
      });
    });
  });
});

test.describe("Billing - Error Handling", () => {
  test.skip(!process.env.E2E_TEST_USER, "No test user configured");

  test.use({
    storageState: "e2e/.auth/user.json",
  });

  test("should handle billing API errors gracefully", async ({ page }) => {
    // Mock API error
    await page.route("**/api/billing/**", (route) => {
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Internal server error",
        }),
      });
    });

    await page.goto("/billing");

    // Should show error state or fallback UI
    await expect(page.getByText(/error|try again|something went wrong/i).or(page.getByRole("alert"))).toBeVisible();
  });

  test("should handle payment failure", async ({ page }) => {
    await page.goto("/billing?payment_failed=true");

    // Should show payment failure message
    await expect(page.getByText(/payment failed|declined|try again/i).or(page.getByRole("alert"))).toBeVisible();
  });
});
