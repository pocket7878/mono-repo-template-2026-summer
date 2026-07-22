import { expect, test } from "@playwright/test";

test("creates and completes a todo", async ({ page }) => {
  await page.goto("/todos");
  await page.getByLabel("Title").fill("Buy milk");
  await page.getByRole("button", { name: "Add todo" }).click();
  await page.getByRole("checkbox", { name: "Buy milk" }).click();
  await expect(page.getByText("Completed")).toBeVisible();
});
