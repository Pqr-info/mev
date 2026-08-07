#!/usr/bin/env bash
set -euo pipefail

echo "=== MAX Inference Bootstrap (WSL Alpine) ==="

GEMINI_BRAIN_DIR_WIN="C:\\Users\\theal\\.gemini"
GEMINI_BRAIN_DIR_WSL="/mnt/c/Users/theal/.gemini"
VLLM_ENV_DIR="/opt/max/vllm-env"
MODELS_DIR="/opt/max/models"
VLLM_PORT="${MAX_VLLM_PORT:-8000}"

# 1. Ensure brain directory is visible in WSL
echo "Checking brain directory at ${GEMINI_BRAIN_DIR_WSL} ..."
if [ ! -d "${GEMINI_BRAIN_DIR_WSL}" ]; then
  echo "ERROR: Brain directory not visible at ${GEMINI_BRAIN_DIR_WSL}"
  exit 1
fi

# 2. Create vLLM virtual environment
echo "Creating vLLM environment at ${VLLM_ENV_DIR} ..."
rm -rf "${VLLM_ENV_DIR}"
python3 -m venv "${VLLM_ENV_DIR}"
source "${VLLM_ENV_DIR}/bin/activate"

# 3. Install vLLM + CUDA deps (assumes CUDA already configured for WSL)
echo "Installing vLLM and dependencies..."
pip install --upgrade pip
pip install vllm==0.6.3.post1 --extra-index-url https://download.pytorch.org/whl/cu121
pip install "transformers==4.44.2" "accelerate" "sentencepiece"

mkdir -p "${MODELS_DIR}"

# 4. Download / prepare Qwen2.5-Coder-32B
echo "Preparing Qwen2.5-Coder-32B ..."
QWEN_MODEL_NAME="Qwen/Qwen2.5-Coder-32B"
QWEN_LOCAL_DIR="${MODELS_DIR}/qwen2.5-coder-32b"

python - <<PY
from huggingface_hub import snapshot_download
snapshot_download(repo_id="${QWEN_MODEL_NAME}", local_dir="${QWEN_LOCAL_DIR}", local_dir_use_symlinks=False)
PY

# 5. Prepare Gemma-4-e4b (GGUF)
echo "Preparing Gemma-4-e4b ..."
GEMMA_LOCAL_DIR="${MODELS_DIR}/gemma-4-e4b"
mkdir -p "${GEMMA_LOCAL_DIR}"
# Move the copied GGUF model into the models directory
if [ ! -f "${GEMMA_LOCAL_DIR}/gemma-4-E4B-it-Q4_K_M.gguf" ] && [ -f "/mnt/c/temp/gemma-4-E4B-it-Q4_K_M.gguf" ]; then
    sudo cp "/mnt/c/temp/gemma-4-E4B-it-Q4_K_M.gguf" "${GEMMA_LOCAL_DIR}/"
    sudo chown $USER:$USER "${GEMMA_LOCAL_DIR}/gemma-4-E4B-it-Q4_K_M.gguf"
fi

# 6. Start vLLM server with both models (multi-model mode)
echo "Starting vLLM server on port ${VLLM_PORT} ..."
cat > /opt/max/vllm_server.py <<'PY'
from vllm import LLM, SamplingParams
from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn
import os

app = FastAPI()

llm_qwen = LLM(model="/opt/max/models/qwen2.5-coder-32b", tensor_parallel_size=1)
llm_gemma = LLM(model="/opt/max/models/gemma-4-e4b/gemma-4-E4B-it-Q4_K_M.gguf", tensor_parallel_size=1)

class InferenceRequest(BaseModel):
    model: str
    prompt: str
    max_tokens: int = 512

@app.post("/infer")
def infer(req: InferenceRequest):
    sp = SamplingParams(max_tokens=req.max_tokens)
    if req.model == "qwen2.5-coder-32b":
        outputs = llm_qwen.generate(req.prompt, sp)
    elif req.model == "gemma-4-e4b":
        outputs = llm_gemma.generate(req.prompt, sp)
    else:
        return {"error": "unknown model"}
    return {"text": outputs[0].outputs[0].text}

@app.get("/health")
def health():
    # Check if models are loaded
    models_ok = (llm_qwen is not None) and (llm_gemma is not None)
    
    # Check brain mount
    brain_path = "/mnt/c/Users/theal/.gemini/transcript.jsonl"
    brain_ok = os.path.exists(brain_path)
    
    status = "ok" if (models_ok and brain_ok) else "degraded"
    return {
        "status": status,
        "models_loaded": models_ok,
        "brain_mounted": brain_ok
    }

@app.get("/state/manifest")
def state_manifest():
    return {
        "snapshot": {
            "id": "SNAP_ID",
            "timestamp": "SNAP_TIMESTAMP"
        },
        "wal": {
            "start_id": "WAL_START_ID",
            "end_id": "WAL_END_ID",
            "events": []
        },
        "models": [
            {
                "name": "qwen2.5-coder-32b",
                "version": "1.0.0",
                "quant": "bf16",
                "path_hash": "..."
            },
            {
                "name": "gemma-4-e4b",
                "version": "1.0.0",
                "quant": "4bit",
                "path_hash": "..."
            }
        ],
        "services": [
            {
                "name": "vllm",
                "version": "0.6.3.post1",
                "config_hash": "...",
                "status": "running"
            }
        ],
        "brain": {
            "path_hash": "...",
            "transcript_hash": "..."
        }
    }

if __name__ == "__main__":
    port = int(os.environ.get("MAX_VLLM_PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
PY

cat > /etc/systemd/system/max-vllm.service <<EOF
[Unit]
Description=MAX vLLM Inference Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/max
Environment="LD_LIBRARY_PATH=/usr/lib/wsl/lib"
ExecStart=/opt/max/vllm-env/bin/python /opt/max/vllm_server.py
StandardOutput=append:/opt/max/vllm_server.log
StandardError=append:/opt/max/vllm_server.log
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable max-vllm
systemctl restart max-vllm

VLLM_PID=$(systemctl show -p MainPID --value max-vllm)
# 7. Health check: read transcript.jsonl
TRANSCRIPT_PATH="${GEMINI_BRAIN_DIR_WSL}/transcript.jsonl"
echo "Checking transcript at ${TRANSCRIPT_PATH} ..."
if [ -f "${TRANSCRIPT_PATH}" ]; then
  head -n 3 "${TRANSCRIPT_PATH}" || true
  echo "Transcript visible."
else
  echo "WARNING: transcript.jsonl not found at ${TRANSCRIPT_PATH}"
fi

echo "vLLM server running (PID=${VLLM_PID}) on port ${VLLM_PORT}"
echo "=== MAX Inference Bootstrap (WSL Alpine) complete ==="
