import { LinearClient } from "@linear/sdk";
import { Hono } from "hono";

const app = new Hono();
const linearClient = new LinearClient({
  apiKey: process.env.LINEAR_API_KEY || "",
});

app.get("/", (c) => {
  return c.text("Hello, World!");
});

app.post("/sync-issues", async (c) => {
  const issuesConn = await linearClient.issues();
  const issues = issuesConn.nodes;

  await Promise.all(
    issues.map(async (issue) => {
      await linearClient.updateIssue(issue.id, {
        addedLabelIds: [issue.id],
      });
    }),
  );

  return c.json(issues);
});

export default app;
