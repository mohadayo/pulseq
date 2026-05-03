import os
import logging
import time
import uuid
from flask import Flask, jsonify, request

app = Flask(__name__)

logging.basicConfig(
    level=os.environ.get("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("pulseq.gateway")

queues: dict[str, list[dict]] = {}


@app.route("/health")
def health():
    return jsonify({"status": "healthy", "service": "gateway", "timestamp": time.time()})


@app.route("/queues", methods=["POST"])
def create_queue():
    data = request.get_json()
    if not data or "name" not in data:
        logger.warning("Queue creation failed: missing name")
        return jsonify({"error": "Queue name is required"}), 400
    name = data["name"]
    if name in queues:
        logger.warning(f"Queue creation failed: '{name}' already exists")
        return jsonify({"error": f"Queue '{name}' already exists"}), 409
    queues[name] = []
    logger.info(f"Queue created: {name}")
    return jsonify({"name": name, "message_count": 0}), 201


@app.route("/queues", methods=["GET"])
def list_queues():
    result = [{"name": k, "message_count": len(v)} for k, v in queues.items()]
    return jsonify(result)


@app.route("/queues/<name>/messages", methods=["POST"])
def publish_message(name: str):
    if name not in queues:
        return jsonify({"error": f"Queue '{name}' not found"}), 404
    data = request.get_json()
    if not data or "body" not in data:
        return jsonify({"error": "Message body is required"}), 400
    message = {
        "id": str(uuid.uuid4()),
        "body": data["body"],
        "published_at": time.time(),
        "status": "pending",
    }
    queues[name].append(message)
    logger.info(f"Message published to '{name}': {message['id']}")
    return jsonify(message), 201


@app.route("/queues/<name>/messages", methods=["GET"])
def get_messages(name: str):
    if name not in queues:
        return jsonify({"error": f"Queue '{name}' not found"}), 404
    return jsonify(queues[name])


@app.route("/queues/<name>/stats", methods=["GET"])
def queue_stats(name: str):
    if name not in queues:
        return jsonify({"error": f"Queue '{name}' not found"}), 404
    messages = queues[name]
    pending = sum(1 for m in messages if m["status"] == "pending")
    processed = sum(1 for m in messages if m["status"] == "processed")
    return jsonify({
        "name": name,
        "total": len(messages),
        "pending": pending,
        "processed": processed,
    })


if __name__ == "__main__":
    port = int(os.environ.get("GATEWAY_PORT", "5000"))
    logger.info(f"Starting PulseQ Gateway on port {port}")
    app.run(host="0.0.0.0", port=port)
