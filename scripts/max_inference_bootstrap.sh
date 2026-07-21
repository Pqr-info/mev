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
python3 -m venv "${VLLM_ENV_DIR}"
source "${VLLM_ENV_DIR}/bin/activate"

# 3. Install vLLM + CUDA deps (assumes CUDA already configured for WSL)
echo "Installing vLLM and dependencies..."
pip install --upgrade pip
pip install vllm "torch>=2.3" "transformers>=4.40" "accelerate" "sentencepiece"

mkdir -p "${MODELS_DIR}"

# 4. Download / prepare Qwen3-Coder-30B
echo "Preparing Qwen3-Coder-30B ..."
QWEN_MODEL_NAME="Qwen/Qwen3-Coder-30B"
QWEN_LOCAL_DIR="${MODELS_DIR}/qwen3-coder-30b"

python - <<PY
from huggingface_hub import snapshot_download
snapshot_download(repo_id="${QWEN_MODEL_NAME}", local_dir="${QWEN_LOCAL_DIR}", local_dir_use_symlinks=False)
PY

# 5. Download / prepare Gemma-4-e4b
echo "Preparing Gemma-4-e4b ..."
GEMMA_MODEL_NAME="google/gemma-4-9b-it-e4b"
GEMMA_LOCAL_DIR="${MODELS_DIR}/gemma-4-e4b"

python - <<PY
from huggingface_hub import snapshot_download
snapshot_download(repo_id="${GEMMA_MODEL_NAME}", local_dir="${GEMMA_LOCAL_DIR}", local_dir_use_symlinks=False)
PY

# 6. Start vLLM server with both models (multi-model mode)
echo "Starting vLLM server on port ${VLLM_PORT} ..."
cat > /opt/max/vllm_server.py <<'PY'
from vllm import LLM, SamplingParams
from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn
import os

app = FastAPI()

llm_qwen = LLM(model="/opt/max/models/qwen3-coder-30b", tensor_parallel_size=1)
llm_gemma = LLM(model="/opt/max/models/gemma-4-e4b", tensor_parallel_size=1)

class InferenceRequest(BaseModel):
    model: str
    prompt: str
    max_tokens: int = 512

@app.post("/infer")
def infer(req: InferenceRequest):
    sp = SamplingParams(max_tokens=req.max_tokens)
    if req.model == "qwen3-coder-30b":
        outputs = llm_qwen.generate(req.prompt, sp)
    elif req.model == "gemma-4-e4b":
        outputs = llm_gemma.generate(req.prompt, sp)
    else:
        return {"error": "unknown model"}
    return {"text": outputs[0].outputs[0].text}

if __name__ == "__main__":
    port = int(os.environ.get("MAX_VLLM_PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
PY

python /opt/max/vllm_server.py &
VLLM_PID=$!

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
