import request from "supertest";
import { app, metrics } from "./index";

beforeEach(() => {
  metrics.length = 0;
});

describe("GET /health", () => {
  it("returns healthy status", async () => {
    const res = await request(app).get("/health");
    expect(res.status).toBe(200);
    expect(res.body.status).toBe("healthy");
    expect(res.body.service).toBe("dashboard");
  });
});

describe("POST /metrics", () => {
  it("records a metric", async () => {
    const res = await request(app)
      .post("/metrics")
      .send({ name: "queue.depth", value: 42, labels: { queue: "orders" } });
    expect(res.status).toBe(201);
    expect(res.body.name).toBe("queue.depth");
    expect(res.body.value).toBe(42);
  });

  it("rejects missing name", async () => {
    const res = await request(app).post("/metrics").send({ value: 10 });
    expect(res.status).toBe(400);
  });

  it("rejects missing value", async () => {
    const res = await request(app).post("/metrics").send({ name: "test" });
    expect(res.status).toBe(400);
  });
});

describe("GET /metrics", () => {
  it("returns all metrics", async () => {
    await request(app).post("/metrics").send({ name: "m1", value: 1 });
    await request(app).post("/metrics").send({ name: "m2", value: 2 });
    const res = await request(app).get("/metrics");
    expect(res.status).toBe(200);
    expect(res.body).toHaveLength(2);
  });
});

describe("GET /metrics/summary", () => {
  it("returns aggregated summary", async () => {
    await request(app).post("/metrics").send({ name: "latency", value: 10 });
    await request(app).post("/metrics").send({ name: "latency", value: 20 });
    await request(app).post("/metrics").send({ name: "latency", value: 30 });
    const res = await request(app).get("/metrics/summary");
    expect(res.status).toBe(200);
    expect(res.body.latency.count).toBe(3);
    expect(res.body.latency.avg).toBe(20);
    expect(res.body.latency.min).toBe(10);
    expect(res.body.latency.max).toBe(30);
  });
});

describe("DELETE /metrics", () => {
  it("clears all metrics", async () => {
    await request(app).post("/metrics").send({ name: "m1", value: 1 });
    const res = await request(app).delete("/metrics");
    expect(res.status).toBe(200);
    const list = await request(app).get("/metrics");
    expect(list.body).toHaveLength(0);
  });
});
