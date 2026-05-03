import express, { Request, Response } from "express";

const app = express();
app.use(express.json());

interface MetricPoint {
  name: string;
  value: number;
  timestamp: number;
  labels: Record<string, string>;
}

const metrics: MetricPoint[] = [];

app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "dashboard",
    timestamp: Date.now() / 1000,
  });
});

app.post("/metrics", (req: Request, res: Response) => {
  const { name, value, labels } = req.body;
  if (!name || value === undefined) {
    res.status(400).json({ error: "name and value are required" });
    return;
  }
  const point: MetricPoint = {
    name,
    value: Number(value),
    timestamp: Date.now() / 1000,
    labels: labels || {},
  };
  metrics.push(point);
  console.log(`[INFO] Metric recorded: ${name}=${value}`);
  res.status(201).json(point);
});

app.get("/metrics", (_req: Request, res: Response) => {
  res.json(metrics);
});

app.get("/metrics/summary", (_req: Request, res: Response) => {
  const summary: Record<string, { count: number; sum: number; avg: number; min: number; max: number }> = {};
  for (const m of metrics) {
    if (!summary[m.name]) {
      summary[m.name] = { count: 0, sum: 0, avg: 0, min: Infinity, max: -Infinity };
    }
    const s = summary[m.name];
    s.count++;
    s.sum += m.value;
    s.min = Math.min(s.min, m.value);
    s.max = Math.max(s.max, m.value);
    s.avg = s.sum / s.count;
  }
  res.json(summary);
});

app.delete("/metrics", (_req: Request, res: Response) => {
  metrics.length = 0;
  console.log("[INFO] All metrics cleared");
  res.json({ message: "All metrics cleared" });
});

const port = parseInt(process.env.DASHBOARD_PORT || "5002", 10);

if (process.env.NODE_ENV !== "test") {
  app.listen(port, "0.0.0.0", () => {
    console.log(`[INFO] Starting PulseQ Dashboard on port ${port}`);
  });
}

export { app, metrics };
