import time
import uuid
from flask import Flask, request, jsonify
from google import genai

import os

app = Flask(__name__)

# Basic client with fallback check
api_key = os.environ.get("GEMINI_API_KEY", "").strip()
client = None
if api_key and api_key != "dummy" and not api_key.startswith("${"):
    try:
        client = genai.Client()
    except Exception as e:
        print(f"Warning: Failed to initialize genai Client: {e}")

def generate_with_gemini(prompt: str):
    start = time.time()
    if client is not None:
        resp = client.models.generate_content(
            model="gemini-3.5-flash",
            contents=prompt,
        )
        output_text = resp.text if hasattr(resp, "text") else str(resp)
    else:
        output_text = f"[MOCK_GEMINI_3.5_FLASH]: Fallback execution for prompt: {prompt[:60]}..."
    
    latency_ms = int((time.time() - start) * 1000)

    return {
        "agent": "gemini",
        "model": "gemini-3.5-flash",
        "request_id": f"req-{uuid.uuid4().hex[:8]}",
        "output": output_text,
        "tokens_in": 12,
        "tokens_out": 24,
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
