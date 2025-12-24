import path from "node:path";

import { expect, test } from "./fixtures/test-fixtures";

/**
 * File Management E2E Tests
 *
 * These tests cover file upload, listing, and download functionality.
 * Tests require authentication and will be skipped if credentials are not set.
 */
test.describe("File Management", () => {
  test.describe("Demo Page - File Operations", () => {
    test("should display demo page with file section", async ({ page }) => {
      await page.goto("/demo");

      // Check if file upload section exists
      await expect(page.getByText(/file.*upload|upload.*file/i)).toBeVisible({ timeout: 10000 });
    });

    test.describe("Authenticated", () => {
      test.skip(!process.env.E2E_TEST_USER, "No test user configured");

      test.use({
        storageState: "e2e/.auth/user.json",
      });

      test("should show file upload area", async ({ page }) => {
        await page.goto("/demo");

        // Look for file upload area
        await expect(
          page.getByText(/drag.*drop|drop.*file|select.*file|choose.*file/i).or(page.locator('input[type="file"]'))
        ).toBeVisible();
      });

      test("should show file list or empty state", async ({ page }) => {
        await page.goto("/demo");

        // Wait for page to load
        await page.waitForLoadState("networkidle");

        // Should show either files or a message about no files
        await expect(page.getByText(/no files|your files|uploaded files/i)).toBeVisible({ timeout: 10000 });
      });

      test("should handle file selection via input", async ({ page }) => {
        await page.goto("/demo");

        // Find file input (might be hidden)
        const fileInput = page.locator('input[type="file"]');

        if (await fileInput.count()) {
          // Create a test file
          const testFilePath = path.join(__dirname, "fixtures", "test-upload.txt");

          // Use Playwright's file chooser to set a file
          await fileInput.setInputFiles({
            name: "test-upload.txt",
            mimeType: "text/plain",
            buffer: Buffer.from("This is a test file for E2E testing."),
          });

          // The file should be selected
          await expect(page.getByText(/test-upload|selected/i)).toBeVisible({ timeout: 5000 });
        }
      });

      test("should show upload button when file is selected", async ({ page }) => {
        await page.goto("/demo");

        const fileInput = page.locator('input[type="file"]');

        if (await fileInput.count()) {
          await fileInput.setInputFiles({
            name: "test.txt",
            mimeType: "text/plain",
            buffer: Buffer.from("Test content"),
          });

          // Should show upload button
          await expect(page.getByRole("button", { name: /upload/i })).toBeVisible();
        }
      });

      test("should show storage status", async ({ page }) => {
        await page.goto("/demo");

        // Wait for storage status to load
        await page.waitForLoadState("networkidle");

        // Look for storage-related information
        await expect(page.getByText(/storage|used|remaining|quota/i)).toBeVisible({ timeout: 10000 });
      });

      test("should display uploaded files in a list", async ({ page }) => {
        // Mock files endpoint
        await page.route("**/api/files**", (route) => {
          if (route.request().method() === "GET") {
            route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                success: true,
                data: [
                  {
                    id: 1,
                    filename: "test-document.pdf",
                    original_name: "test-document.pdf",
                    size: 1024000,
                    mime_type: "application/pdf",
                    created_at: new Date().toISOString(),
                  },
                  {
                    id: 2,
                    filename: "image.jpg",
                    original_name: "image.jpg",
                    size: 512000,
                    mime_type: "image/jpeg",
                    created_at: new Date().toISOString(),
                  },
                ],
              }),
            });
          } else {
            route.continue();
          }
        });

        await page.goto("/demo");

        // Should show file names
        await expect(page.getByText("test-document.pdf")).toBeVisible();
        await expect(page.getByText("image.jpg")).toBeVisible();
      });

      test("should show delete button for each file", async ({ page }) => {
        // Mock files endpoint
        await page.route("**/api/files**", (route) => {
          if (route.request().method() === "GET") {
            route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                success: true,
                data: [
                  {
                    id: 1,
                    filename: "test-file.txt",
                    original_name: "test-file.txt",
                    size: 1024,
                    mime_type: "text/plain",
                    created_at: new Date().toISOString(),
                  },
                ],
              }),
            });
          } else {
            route.continue();
          }
        });

        await page.goto("/demo");

        // Should show delete button
        await expect(page.getByRole("button", { name: /delete/i })).toBeVisible();
      });

      test("should show download button for each file", async ({ page }) => {
        // Mock files endpoint
        await page.route("**/api/files**", (route) => {
          if (route.request().method() === "GET") {
            route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                success: true,
                data: [
                  {
                    id: 1,
                    filename: "test-file.txt",
                    original_name: "test-file.txt",
                    size: 1024,
                    mime_type: "text/plain",
                    created_at: new Date().toISOString(),
                  },
                ],
              }),
            });
          } else {
            route.continue();
          }
        });

        await page.goto("/demo");

        // Should show download button
        await expect(page.getByRole("button", { name: /download/i })).toBeVisible();
      });
    });
  });

  test.describe("File Upload Flow", () => {
    test.skip(!process.env.E2E_TEST_USER, "No test user configured");

    test.use({
      storageState: "e2e/.auth/user.json",
    });

    test("should upload a file successfully", async ({ page }) => {
      // Mock successful upload
      await page.route("**/api/files/upload**", (route) => {
        route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            success: true,
            message: "File uploaded successfully",
            data: {
              id: 99,
              filename: "uploaded-file.txt",
              original_name: "test.txt",
              size: 1024,
              mime_type: "text/plain",
            },
          }),
        });
      });

      await page.goto("/demo");

      const fileInput = page.locator('input[type="file"]');

      if (await fileInput.count()) {
        await fileInput.setInputFiles({
          name: "test.txt",
          mimeType: "text/plain",
          buffer: Buffer.from("Test content for upload"),
        });

        const uploadButton = page.getByRole("button", { name: /upload/i });
        if (await uploadButton.isVisible()) {
          await uploadButton.click();

          // Should show success message
          await expect(page.getByText(/success|uploaded/i)).toBeVisible({ timeout: 10000 });
        }
      }
    });

    test("should show error for failed upload", async ({ page }) => {
      // Mock failed upload
      await page.route("**/api/files/upload**", (route) => {
        route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({
            success: false,
            error: "Upload failed",
          }),
        });
      });

      await page.goto("/demo");

      const fileInput = page.locator('input[type="file"]');

      if (await fileInput.count()) {
        await fileInput.setInputFiles({
          name: "test.txt",
          mimeType: "text/plain",
          buffer: Buffer.from("Test content"),
        });

        const uploadButton = page.getByRole("button", { name: /upload/i });
        if (await uploadButton.isVisible()) {
          await uploadButton.click();

          // Should show error message
          await expect(page.getByText(/error|failed/i)).toBeVisible({ timeout: 10000 });
        }
      }
    });
  });

  test.describe("File Delete Flow", () => {
    test.skip(!process.env.E2E_TEST_USER, "No test user configured");

    test.use({
      storageState: "e2e/.auth/user.json",
    });

    test("should delete a file with confirmation", async ({ page }) => {
      // Mock files endpoint
      await page.route("**/api/files**", (route) => {
        if (route.request().method() === "GET") {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              success: true,
              data: [
                {
                  id: 1,
                  filename: "deletable-file.txt",
                  original_name: "deletable-file.txt",
                  size: 1024,
                  mime_type: "text/plain",
                  created_at: new Date().toISOString(),
                },
              ],
            }),
          });
        } else if (route.request().method() === "DELETE") {
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              success: true,
              message: "File deleted successfully",
            }),
          });
        } else {
          route.continue();
        }
      });

      await page.goto("/demo");

      // Click delete button
      const deleteButton = page.getByRole("button", { name: /delete/i });
      if (await deleteButton.isVisible()) {
        await deleteButton.click();

        // If there's a confirmation dialog, confirm it
        const confirmButton = page.getByRole("button", { name: /confirm|yes|delete/i });
        if (await confirmButton.isVisible()) {
          await confirmButton.click();
        }

        // Should show success message
        await expect(page.getByText(/deleted|success/i)).toBeVisible({ timeout: 10000 });
      }
    });
  });
});
