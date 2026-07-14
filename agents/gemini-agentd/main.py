import time
import uuid
from flask import Flask, request, jsonify
from google import genai

app = Flask(__name__)

# Basic client; you can extend with explicit timeouts/retries via config
client = genai.Client()

def generate_with_gemini(prompt: str):
    start = time.time()
    # Adjust to the exact Interactions/GenAI API you’re using
    resp = client.models.generate_content(
        model="gemini-3.5-flash",
        contents=prompt,
    )
    output_text = resp.text if hasattr(resp, "text") else str(resp)
    latency_ms = int((time.time() - start) * 1000)

    # Token counts may depend on API; stubbed here
    return {
        "agent": "gemini",
        "model": "gemini-3.5-flash",
        "request_id": f"req-{uuid.uuid4().hex[:8]}",
        "output": output_text,
        "tokens_in": None,
        "tokens_out": None,
        "latency_ms": latency_ms,
    }

@app.route("/interact", methods=["POST"])
def interact():
    data = request.get_json(force=True, silent=True) or {}
    prompt = data.get("prompt")

    if not isinstance(prompt, str) or not prompt.strip():
        return jsonify({
            "error": "invalid_prompt",
            "message": "prompt must be a non-empty string",
        }), 400

    try:
        envelope = generate_with_gemini(prompt)
        return jsonify(envelope), 200
    except Exception as e:
        return jsonify({
            "error": "gemini_failure",
            "message": str(e),
        }), 502

@app.route("/ping", methods=["GET"])
def ping():
    return jsonify({"status": "ok", "service": "gemini-agentd"}), 200

if __name__ == "__main__":
    # You can tune host/port via env
    app.run(host="0.0.0.0", port=8081)
