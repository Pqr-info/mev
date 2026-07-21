import json
import threading
import time
import uuid
import functools
from datetime import datetime, timezone
from pathlib import Path

import requests
from flask import Flask, request, jsonify

BACKCHANNEL_PATH = Path(".copilot_backchannel.json")
GEMINI_AGENT_URL = "http://gemini-agentd:8081/interact"
HEARTBEAT_INTERVAL_SEC = 10
INTENT_POLL_INTERVAL_SEC = 1
MAX_HEARTBEAT_FAILURES = 3
MAX_RUNNING_INTENT_AGE_SEC = 60

app = Flask(__name__)
lock = threading.Lock()


def now_iso():
    return datetime.now(timezone.utc).isoformat()


def load_state():
    if not BACKCHANNEL_PATH.exists():
        return {
            "version": "bcspine-1",
            "epoch": 0,
            "agents": {
                "gemini": {
                    "status": "online",
                    "model": "gemini-3.5-flash",
                    "heartbeat_at": None,
                    "last_request": None,
                    "last_response": None,
                    "last_error": None,
                }
            },
            "intents": [],
        }
    with BACKCHANNEL_PATH.open("r", encoding="utf-8") as f:
        return json.load(f)


def save_state(state):
    with BACKCHANNEL_PATH.open("w", encoding="utf-8") as f:
        json.dump(state, f, indent=2)


def with_state(fn):
    @functools.wraps(fn)
    def wrapper(*args, **kwargs):
        with lock:
            state = load_state()
            result = fn(state, *args, **kwargs)
            save_state(state)
            return result
    return wrapper


@app.route("/ping", methods=["GET"])
def ping():
    return jsonify({"status": "ok", "service": "bcpd"}), 200


@app.route("/state", methods=["GET"])
@with_state
def state_endpoint(state):
    return jsonify(state), 200


@app.route("/intent", methods=["POST"])
@with_state
def intent_endpoint(state):
    data = request.get_json(force=True, silent=True) or {}
    agent = data.get("agent", "gemini")
    kind = data.get("kind", "analysis")
    payload = data.get("payload") or {}

    intent_id = f"intent-{uuid.uuid4().hex[:6]}"
    intent = {
        "id": intent_id,
        "agent": agent,
        "kind": kind,
        "payload": payload,
        "created_at": now_iso(),
        "status": "pending",
        "error": None,
    }
    state.setdefault("intents", []).append(intent)
    return jsonify({"id": intent_id, "status": "pending"}), 202


@app.route("/proposal", methods=["POST"])
@with_state
def proposal_endpoint(state):
    from datetime import timedelta
    data = request.get_json(force=True, silent=True) or {}
    kind = data.get("kind", "policy_mutation")
    description = data.get("description", "")
    proposer = data.get("proposer", "")
    required_votes = data.get("required_votes", 3)

    proposal_id = f"prop-{uuid.uuid4().hex[:6]}"
    intent_id = f"intent-prop-{uuid.uuid4().hex[:6]}"

    votes = {
        "gemini": "pending",
        "gemma": "pending",
        "copilot": "pending",
        "antigravity": "pending"
    }
    if proposer in votes:
        votes[proposer] = "approve"

    created = datetime.now(timezone.utc)
    expires = created + timedelta(minutes=10)

    intent = {
        "id": intent_id,
        "agent": "mesh-consensus",
        "kind": "proposal",
        "payload": {
            "proposal_id": proposal_id,
            "kind": kind,
            "description": description,
            "proposer": proposer,
            "votes": votes,
            "required_votes": required_votes,
            "result": "pending",
            "created_at": created.isoformat(),
            "expires_at": expires.isoformat()
        },
        "status": "pending",
        "error": None
    }

    state.setdefault("intents", []).append(intent)
    return jsonify({
        "proposal_id": proposal_id,
        "intent_id": intent_id,
        "status": "pending"
    }), 201


@app.route("/proposal/vote", methods=["POST"])
@with_state
def proposal_vote_endpoint(state):
    data = request.get_json(force=True, silent=True) or {}
    proposal_id = data.get("proposal_id")
    voter = data.get("voter")
    vote = data.get("vote")  # "approve" | "reject" | "abstain"

    if not proposal_id or not voter or not vote:
        return jsonify({"error": "missing_fields"}), 400

    # Find proposal
    proposal_intent = None
    for intent in state.get("intents", []):
        if (intent.get("agent") == "mesh-consensus" and
            intent.get("kind") == "proposal" and
            intent.get("payload", {}).get("proposal_id") == proposal_id):
            proposal_intent = intent
            break

    if not proposal_intent:
        return jsonify({"error": "proposal_not_found"}), 404

    payload = proposal_intent.setdefault("payload", {})
    if payload.get("result") != "pending":
        return jsonify({
            "error": "proposal_closed",
            "message": f"Proposal is already finalized as: {payload.get('result')}"
        }), 400

    votes = payload.setdefault("votes", {})
    votes[voter] = vote

    # Evaluate consensus
    now = datetime.now(timezone.utc)
    expires_at = datetime.fromisoformat(payload["expires_at"])
    
    if now > expires_at:
        payload["result"] = "expired"
        proposal_intent["status"] = "done"
    else:
        approve_count = sum(1 for v in votes.values() if v == "approve")
        reject_count = sum(1 for v in votes.values() if v == "reject")
        total_voters = len(votes) # 4

        if approve_count >= payload["required_votes"]:
            payload["result"] = "approved"
            proposal_intent["status"] = "done"
        elif total_voters - reject_count < payload["required_votes"]:
            payload["result"] = "rejected"
            proposal_intent["status"] = "done"

    return jsonify({
        "proposal_id": proposal_id,
        "result": payload["result"],
        "votes": votes
    }), 200


@app.route("/lineage", methods=["GET"])
@with_state
def lineage_endpoint(state):
    lineage = state.get("lineage", [])
    sorted_lineage = sorted(
        lineage,
        key=lambda x: (x.get("epoch", 0), x.get("timestamp", ""))
    )
    return jsonify({"lineage": sorted_lineage}), 200


@app.route("/proposal/rollback", methods=["POST"])
@with_state
def proposal_rollback_endpoint(state):
    from datetime import timedelta
    data = request.get_json(force=True, silent=True) or {}
    lineage_id = data.get("lineage_id")
    proposer = data.get("proposer", "")
    required_votes = data.get("required_votes", 3)

    if not lineage_id:
        return jsonify({"error": "missing_lineage_id"}), 400

    # Locate lineage event
    lineage_event = None
    for entry in state.get("lineage", []):
        if entry.get("lineage_id") == lineage_id:
            lineage_event = entry
            break

    if not lineage_event:
        return jsonify({"error": "lineage_event_not_found"}), 404

    mutation = lineage_event.get("mutation", {})
    target_file = mutation.get("file")
    target_key = mutation.get("key")
    target_value = mutation.get("prev_value")

    proposal_id = f"prop-{uuid.uuid4().hex[:6]}"
    intent_id = f"intent-prop-{uuid.uuid4().hex[:6]}"

    votes = {
        "gemini": "pending",
        "gemma": "pending",
        "copilot": "pending",
        "antigravity": "pending"
    }
    if proposer in votes:
        votes[proposer] = "approve"

    created = datetime.now(timezone.utc)
    expires = created + timedelta(minutes=10)

    intent = {
        "id": intent_id,
        "agent": "mesh-consensus",
        "kind": "proposal",
        "payload": {
            "proposal_id": proposal_id,
            "kind": "rollback",
            "description": f"Rollback of lineage event {lineage_id}",
            "proposer": proposer,
            "votes": votes,
            "required_votes": required_votes,
            "result": "pending",
            "created_at": created.isoformat(),
            "expires_at": expires.isoformat(),
            "mutation_details": {
                "file": target_file,
                "key": target_key,
                "value": target_value
            }
        },
        "status": "pending",
        "error": None
    }

    state.setdefault("intents", []).append(intent)
    return jsonify({
        "proposal_id": proposal_id,
        "intent_id": intent_id,
        "status": "pending"
    }), 201


def execute_simulation(intent, state):
    payload = intent.get("payload", {})
    proposal = payload.get("proposal", {})
    mut = proposal.get("mutation_details", {})
    file_name = mut.get("file")
    key_path = mut.get("key")
    candidate_value = mut.get("value")

    lineage = state.get("lineage", [])
    relevant_events = []
    for entry in lineage:
        m = entry.get("mutation", {})
        if m.get("file") == file_name and m.get("key") == key_path:
            relevant_events.append(entry)

    recent_events = relevant_events[-10:]
    history_str = "\n".join([
        f"- Time: {e.get('timestamp')}, Proposer: {e.get('proposer')}, Value: {e.get('mutation', {}).get('prev_value')} -> {e.get('mutation', {}).get('new_value')}"
        for e in recent_events
    ])

    prompt = f"""
Governance simulation requested.
Target config: '{file_name}', key: '{key_path}'.
Hypothetical new value: {candidate_value}

Here is the recent update history for this key:
{history_str if history_str else "No recent history."}

Predict the outcome of proposing this mutation to the mesh:
1. Expected votes from: gemini, gemma, copilot, antigravity. Options: approve, reject, abstain, pending.
2. Stability score (0.0 to 1.0, where 1.0 is highly stable/low risk).
3. Rollback risk (low, medium, high).
4. Recommended value (if you believe the candidate value is unstable).
5. Explanatory notes.

You must respond in JSON format matching the following schema exactly:
{{
  "predicted_votes": {{
    "gemini": "approve/reject/abstain",
    "gemma": "approve/reject/abstain",
    "copilot": "approve/reject/abstain",
    "antigravity": "approve/reject/abstain"
  }},
  "stability_score": 0.85,
  "rollback_risk": "low/medium/high",
  "recommended_value": 30.0,
  "notes": "explanatory notes"
}}
"""
    try:
        resp = requests.post(
            GEMINI_AGENT_URL,
            json={"prompt": prompt},
            timeout=30
        )
        if resp.status_code == 200:
            result = resp.json()
            output_text = result.get("output", "")
            if "```json" in output_text:
                output_text = output_text.split("```json")[1].split("```")[0].strip()
            elif "```" in output_text:
                output_text = output_text.split("```")[1].split("```")[0].strip()
            
            analysis = json.loads(output_text.strip())
            payload["result"] = analysis
            intent["status"] = "done"
            intent["error"] = None
        else:
            intent["status"] = "error"
            intent["error"] = {"message": f"gemini-agentd returned status {resp.status_code}", "code": "gemini_error"}
    except Exception as e:
        intent["status"] = "error"
        intent["error"] = {"message": str(e), "code": "simulation_failed"}


@app.route("/simulation", methods=["POST"])
@with_state
def simulation_endpoint(state):
    data = request.get_json(force=True, silent=True) or {}
    proposal = data.get("proposal", {})
    context = data.get("context", {"window_minutes": 30, "max_lineage_events": 50})

    if not proposal or "mutation_details" not in proposal:
        return jsonify({"error": "missing_proposal_details"}), 400

    sim_id = f"sim-{uuid.uuid4().hex[:6]}"
    intent_id = f"intent-sim-{uuid.uuid4().hex[:6]}"

    intent = {
        "id": intent_id,
        "agent": "mesh-consensus",
        "kind": "simulation",
        "payload": {
            "simulation_id": sim_id,
            "proposal": proposal,
            "context": context,
            "result": None
        },
        "created_at": datetime.now(timezone.utc).isoformat(),
        "status": "pending",
        "error": None
    }

    execute_simulation(intent, state)
    state.setdefault("intents", []).append(intent)

    if intent["status"] == "error":
        return jsonify({"error": "simulation_failed", "message": intent["error"]["message"]}), 500

    return jsonify(intent["payload"]["result"]), 200


@with_state
def process_pending_intents(state):
    now = time.time()
    for intent in state.get("intents", []):
        status = intent.get("status")
        if status == "running":
            created = datetime.fromisoformat(intent["created_at"]).timestamp()
            if now - created > MAX_RUNNING_INTENT_AGE_SEC:
                intent["status"] = "error"
                intent["error"] = {
                    "message": "intent exceeded max running age",
                    "code": "timeout",
                }
            continue

        if status != "pending":
            continue

        if intent.get("agent") != "gemini":
            continue  # future: route to other agents

        # Mark running
        intent["status"] = "running"
        payload = intent.get("payload") or {}
        prompt = payload.get("prompt", "")

        try:
            resp = requests.post(
                GEMINI_AGENT_URL,
                json={"prompt": prompt},
                timeout=30,
            )
            if resp.status_code == 200:
                envelope = resp.json()
                state["agents"]["gemini"]["last_request"] = prompt
                state["agents"]["gemini"]["last_response"] = envelope
                state["agents"]["gemini"]["last_error"] = None
                intent["status"] = "done"
            else:
                intent["status"] = "error"
                intent["error"] = {
                    "message": f"gemini-agentd HTTP {resp.status_code}",
                    "code": "gemini_http_error",
                }
                state["agents"]["gemini"]["last_error"] = intent["error"]
        except Exception as e:
            intent["status"] = "error"
            intent["error"] = {
                "message": str(e),
                "code": "gemini_exception",
            }
            state["agents"]["gemini"]["last_error"] = intent["error"]


heartbeat_failures = 0


@with_state
def heartbeat_tick(state):
    global heartbeat_failures
    try:
        url = GEMINI_AGENT_URL.replace("/interact", "/ping")
        resp = requests.get(url, timeout=5)
        if resp.status_code == 200:
            state["agents"]["gemini"]["heartbeat_at"] = now_iso()
            if state["agents"]["gemini"]["status"] in ("offline", "degraded"):
                state["agents"]["gemini"]["status"] = "online"
            heartbeat_failures = 0
            return
    except Exception:
        pass

    heartbeat_failures += 1
    if heartbeat_failures >= MAX_HEARTBEAT_FAILURES:
        state["agents"]["gemini"]["status"] = "degraded"


def set_nested_value(d, path, value):
    keys = path.split('.')
    current = d
    for key in keys[:-1]:
        if key not in current or not isinstance(current[key], dict):
            return False, "path_invalid"
        current = current[key]
    
    last_key = keys[-1]
    if last_key not in current:
        return False, "key_not_found"
        
    if current[last_key] == value:
        return True, "idempotent"
        
    current[last_key] = value
    return True, "updated"


def get_nested_value(d, path):
    keys = path.split('.')
    current = d
    for key in keys:
        if not isinstance(current, dict) or key not in current:
            return None
        current = current[key]
    return current


def apply_policy_mutation(mutation_details):
    file_path = Path(mutation_details.get("file", ""))
    key_path = mutation_details.get("key", "")
    value = mutation_details.get("value")

    if not file_path.exists():
        return {"message": f"file not found: {file_path}", "code": "file_not_found"}

    try:
        with file_path.open("r", encoding="utf-8") as f:
            data = json.load(f)
    except Exception as e:
        return {"message": f"failed to load JSON: {str(e)}", "code": "io_error"}

    # Capture prev_value before update
    prev_val = get_nested_value(data, key_path)
    mutation_details["prev_value"] = prev_val

    success, code = set_nested_value(data, key_path, value)
    if not success:
        return {"message": f"nested path {key_path} is invalid", "code": code}

    if code == "idempotent":
        return None

    try:
        with file_path.open("w", encoding="utf-8") as f:
            json.dump(data, f, indent=2)
    except Exception as e:
        return {"message": f"failed to write JSON: {str(e)}", "code": "io_error"}

    return None


@with_state
def policy_application_tick(state):
    for intent in state.get("intents", []):
        if intent.get("agent") != "mesh-consensus" or intent.get("kind") != "proposal":
            continue
        
        payload = intent.get("payload", {})
        if payload.get("result") != "approved" or payload.get("applied") is True:
            continue
            
        mutation_details = payload.get("mutation_details")
        if not mutation_details:
            payload["applied"] = True
            payload["applied_at"] = now_iso()
            continue

        err = apply_policy_mutation(mutation_details)
        if err:
            payload["applied"] = False
            payload["apply_error"] = err
        else:
            payload["applied"] = True
            payload["applied_at"] = now_iso()
            payload["apply_error"] = None
            
            # Generate lineage event
            epoch = state.get("epoch", 0)
            lineage_list = state.setdefault("lineage", [])
            lineage_id = f"lin-{len(lineage_list) + 1:06d}"
            
            lineage_event = {
                "lineage_id": lineage_id,
                "epoch": epoch,
                "proposal_id": payload.get("proposal_id"),
                "kind": payload.get("kind"),
                "proposer": payload.get("proposer"),
                "description": payload.get("description"),
                "mutation": {
                    "file": mutation_details.get("file"),
                    "key": mutation_details.get("key"),
                    "prev_value": mutation_details.get("prev_value"),
                    "new_value": mutation_details.get("value")
                },
                "timestamp": payload["applied_at"]
            }
            lineage_list.append(lineage_event)


COOLDOWNS = {}


@with_state
def drift_detection_tick(state):
    global COOLDOWNS
    lineage = state.get("lineage", [])
    intents = state.get("intents", [])
    if not lineage:
        return

    mutations_by_key = {}
    for entry in lineage:
        mut = entry.get("mutation", {})
        file_name = mut.get("file")
        key_path = mut.get("key")
        if not file_name or not key_path:
            continue
        full_key = f"{file_name}::{key_path}"
        mutations_by_key.setdefault(full_key, []).append(entry)

    now = time.time()

    for full_key, events in mutations_by_key.items():
        if full_key in COOLDOWNS and now < COOLDOWNS[full_key]:
            continue

        # Look up forecast oscillation probability
        oscillation_prob = 0.5
        forecasts = [
            i for i in intents
            if i.get("kind") == "forecast"
            and i.get("status") == "done"
            and i.get("payload", {}).get("dotted_key") == full_key.split("::", 1)[1] # Match key
        ]
        if forecasts:
            forecasts.sort(key=lambda x: x.get("created_at", ""), reverse=True)
            result = forecasts[0].get("payload", {}).get("result") or {}
            oscillation_prob = float(result.get("oscillation_probability", 0.5))

        # Adjust sensitivity and cooldown dynamically
        sensitivity = 3
        cooldown_duration = 300 # Default 5 mins

        if oscillation_prob >= 0.75:
            sensitivity = 2 # Higher sensitivity (fewer mutations needed to flag)
            cooldown_duration = 120 # Shorter cooldown (respond quickly to persistent instability)
        elif oscillation_prob < 0.30:
            sensitivity = 4 # Lower sensitivity
            cooldown_duration = 600 # Longer cooldown (low risk of immediate drift)

        recent_events = []
        for event in events:
            ts = event.get("timestamp")
            if not ts:
                continue
            try:
                event_time = datetime.fromisoformat(ts).timestamp()
                if now - event_time < 600:
                    recent_events.append(event)
            except Exception:
                pass

        if len(recent_events) >= sensitivity:
            history_str = "\n".join([
                f"- Time: {e.get('timestamp')}, Proposer: {e.get('proposer')}, Value: {e.get('mutation', {}).get('prev_value')} -> {e.get('mutation', {}).get('new_value')}"
                for e in recent_events
            ])
            
            file_name, key_path = full_key.split("::", 1)
            prompt = f"""
Parameter drift/oscillation detected for config '{file_name}', key '{key_path}'.
Here is the recent update history (last 10 minutes):
{history_str}

Analyze this pattern. Determine if it represents unstable drift or oscillation.
If it is unstable, propose a stable, corrective target value.
You must respond in JSON format matching the following schema exactly:
{{
  "unstable": true,
  "reason": "explanation of analysis",
  "recommended_value": 30.0
}}
"""
            try:
                resp = requests.post(
                    GEMINI_AGENT_URL,
                    json={"prompt": prompt},
                    timeout=30
                )
                if resp.status_code == 200:
                    result = resp.json()
                    output_text = result.get("output", "")
                    if "```json" in output_text:
                        output_text = output_text.split("```json")[1].split("```")[0].strip()
                    elif "```" in output_text:
                        output_text = output_text.split("```")[1].split("```")[0].strip()
                    
                    analysis = json.loads(output_text.strip())
                    if analysis.get("unstable") is True and analysis.get("recommended_value") is not None:
                        recommended_value = analysis.get("recommended_value")
                        COOLDOWNS[full_key] = now + cooldown_duration
                        
                        proposal_id = f"prop-drift-{uuid.uuid4().hex[:6]}"
                        intent_id = f"intent-prop-{uuid.uuid4().hex[:6]}"
                        
                        created = datetime.now(timezone.utc)
                        from datetime import timedelta
                        expires = created + timedelta(minutes=10)
                        
                        intent = {
                            "id": intent_id,
                            "agent": "mesh-consensus",
                            "kind": "proposal",
                            "payload": {
                                "proposal_id": proposal_id,
                                "kind": "policy_mutation",
                                "description": f"Corrective drift stabilization: {analysis.get('reason')}",
                                "proposer": "gemini",
                                "votes": {
                                    "gemini": "approve",
                                    "gemma": "pending",
                                    "copilot": "pending",
                                    "antigravity": "pending"
                                },
                                "required_votes": 3,
                                "result": "pending",
                                "created_at": created.isoformat(),
                                "expires_at": expires.isoformat(),
                                "mutation_details": {
                                    "file": file_name,
                                    "key": key_path,
                                    "value": recommended_value
                                }
                            },
                            "status": "pending",
                            "error": None
                        }
                        state.setdefault("intents", []).append(intent)
            except Exception as e:
                print(f"Drift detection handler error: {e}")


def execute_forecasting(intent, state):
    payload = intent.get("payload", {})
    key_path = payload.get("dotted_key")
    
    metrics = state.get("metrics", {})
    param_metrics = metrics.get("parameters", {}).get(key_path, {})
    
    lineage = state.get("lineage", [])
    relevant_events = []
    for entry in lineage:
        m = entry.get("mutation", {})
        if m.get("key") == key_path:
            relevant_events.append(entry)

    recent_events = relevant_events[-10:]
    history_str = "\n".join([
        f"- Time: {e.get('timestamp')}, Proposer: {e.get('proposer')}, Value: {e.get('mutation', {}).get('prev_value')} -> {e.get('mutation', {}).get('new_value')}"
        for e in recent_events
    ])

    prompt = f"""
Predictive drift forecasting requested.
Target config parameter: '{key_path}'.

Here is the recent update history for this key:
{history_str if history_str else "No recent history."}

Here are the stability metrics for this parameter:
- Volatility Index: {param_metrics.get('volatility_index', 0.0)}
- Mutations (Last 24h): {param_metrics.get('mutations_last_24h', 0)}
- Rollbacks (Last 24h): {param_metrics.get('rollbacks_last_24h', 0)}
- Average Simulation Stability Score: {param_metrics.get('avg_stability_score', 1.0)}

Forecast future instability/drift windows for this parameter:
1. Oscillation probability (float 0.0 to 1.0, representing risk in the next 6 hours).
2. Predicted drift window (e.g. "next_4h", "next_2h", or "none").
3. Recommended stabilizing target value (number, string, boolean or null).
4. Explanatory notes.

You must respond in JSON format matching the following schema exactly:
{{
  "oscillation_probability": 0.72,
  "predicted_drift_window": "next_4h",
  "recommended_value": 30.0,
  "notes": "explanatory notes"
}}
"""
    try:
        resp = requests.post(
            GEMINI_AGENT_URL,
            json={"prompt": prompt},
            timeout=30
        )
        if resp.status_code == 200:
            result = resp.json()
            output_text = result.get("output", "")
            if "```json" in output_text:
                output_text = output_text.split("```json")[1].split("```")[0].strip()
            elif "```" in output_text:
                output_text = output_text.split("```")[1].split("```")[0].strip()
            
            analysis = json.loads(output_text.strip())
            payload["result"] = analysis
            intent["status"] = "done"
            intent["error"] = None
        else:
            intent["status"] = "error"
            intent["error"] = {"message": f"gemini-agentd returned status {resp.status_code}", "code": "gemini_error"}
    except Exception as e:
        intent["status"] = "error"
        intent["error"] = {"message": str(e), "code": "forecast_failed"}


@app.route("/forecast", methods=["POST"])
@with_state
def forecast_endpoint(state):
    data = request.get_json(force=True, silent=True) or {}
    key_path = data.get("dotted_key")
    context = data.get("context", {"window_hours": 24})

    if not key_path:
        return jsonify({"error": "missing_dotted_key"}), 400

    fc_id = f"fc-{uuid.uuid4().hex[:6]}"
    intent_id = f"intent-fc-{uuid.uuid4().hex[:6]}"

    intent = {
        "id": intent_id,
        "agent": "mesh-consensus",
        "kind": "forecast",
        "payload": {
            "forecast_id": fc_id,
            "dotted_key": key_path,
            "context": context,
            "result": None
        },
        "created_at": datetime.now(timezone.utc).isoformat(),
        "status": "pending",
        "error": None
    }

    execute_forecasting(intent, state)
    state.setdefault("intents", []).append(intent)

    if intent["status"] == "error":
        return jsonify({"error": "forecast_failed", "message": intent["error"]["message"]}), 500

    return jsonify(intent["payload"]["result"]), 200


@app.route("/metrics", methods=["GET"])
@with_state
def metrics_endpoint(state):
    metrics = state.get("metrics", {})
    return jsonify(metrics), 200


@with_state
def metrics_tick(state):
    now = time.time()
    day_ago = now - 86400

    lineage = state.get("lineage", [])
    intents = state.get("intents", [])

    params_metrics = {}

    for entry in lineage:
        mut = entry.get("mutation", {})
        file_name = mut.get("file")
        key_path = mut.get("key")
        if file_name and key_path:
            params_metrics[f"{file_name}::{key_path}"] = {
                "mutations_last_24h": 0,
                "rollbacks_last_24h": 0,
                "drift_flags_last_24h": 0,
                "simulations_run_last_24h": 0,
                "avg_stability_score": 1.0,
                "volatility_index": 0.0,
                "stability_scores": []
            }

    for intent in intents:
        if intent.get("kind") == "simulation":
            payload = intent.get("payload", {})
            prop = payload.get("proposal", {})
            mut = prop.get("mutation_details", {})
            file_name = mut.get("file")
            key_path = mut.get("key")
            if file_name and key_path:
                full_key = f"{file_name}::{key_path}"
                if full_key not in params_metrics:
                    params_metrics[full_key] = {
                        "mutations_last_24h": 0,
                        "rollbacks_last_24h": 0,
                        "drift_flags_last_24h": 0,
                        "simulations_run_last_24h": 0,
                        "avg_stability_score": 1.0,
                        "volatility_index": 0.0,
                        "stability_scores": []
                    }

    total_mutations = 0
    total_rollbacks = 0
    total_drift_flags = 0
    total_simulations = 0

    for entry in lineage:
        ts = entry.get("timestamp")
        if not ts:
            continue
        try:
            event_time = datetime.fromisoformat(ts).timestamp()
            if event_time >= day_ago:
                mut = entry.get("mutation", {})
                file_name = mut.get("file")
                key_path = mut.get("key")
                if file_name and key_path:
                    full_key = f"{file_name}::{key_path}"
                    if entry.get("kind") == "rollback":
                        params_metrics[full_key]["rollbacks_last_24h"] += 1
                        total_rollbacks += 1
                    else:
                        params_metrics[full_key]["mutations_last_24h"] += 1
                        total_mutations += 1
        except Exception:
            pass

    for intent in intents:
        ts = intent.get("created_at")
        if not ts:
            continue
        try:
            event_time = datetime.fromisoformat(ts).timestamp()
            if event_time >= day_ago:
                kind = intent.get("kind")
                payload = intent.get("payload", {})
                
                if kind == "simulation":
                    prop = payload.get("proposal", {})
                    mut = prop.get("mutation_details", {})
                    file_name = mut.get("file")
                    key_path = mut.get("key")
                    if file_name and key_path:
                        full_key = f"{file_name}::{key_path}"
                        params_metrics[full_key]["simulations_run_last_24h"] += 1
                        total_simulations += 1
                        
                        result = payload.get("result")
                        if isinstance(result, dict) and "stability_score" in result:
                            params_metrics[full_key]["stability_scores"].append(float(result["stability_score"]))
                
                elif kind == "proposal":
                    if payload.get("proposer") == "gemini":
                        mut_details = payload.get("mutation_details", {})
                        file_name = mut_details.get("file")
                        key_path = mut_details.get("key")
                        if file_name and key_path:
                            full_key = f"{file_name}::{key_path}"
                            params_metrics[full_key]["drift_flags_last_24h"] += 1
                            total_drift_flags += 1
        except Exception:
            pass

    all_scores = []
    formatted_params = {}
    for full_key, metric in params_metrics.items():
        scores = metric.pop("stability_scores", [])
        if scores:
            avg_score = sum(scores) / len(scores)
            all_scores.extend(scores)
        else:
            avg_score = 1.0
        metric["avg_stability_score"] = round(avg_score, 2)

        muts = metric["mutations_last_24h"]
        rolls = metric["rollbacks_last_24h"]
        drifts = metric["drift_flags_last_24h"]
        vol = (muts + 2 * rolls + drifts) / 4.0
        metric["volatility_index"] = round(min(1.0, vol), 2)
        formatted_params[full_key] = metric

    global_avg_stability = sum(all_scores) / len(all_scores) if all_scores else 1.0
    global_stability = max(0.0, global_avg_stability - (total_rollbacks * 0.1) - (total_drift_flags * 0.05))
    global_stability = round(min(1.0, global_stability), 2)

    health = "stable"
    if global_stability < 0.6:
        health = "unstable"
    elif global_stability < 0.85:
        health = "moderate"

    state["metrics"] = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "parameters": formatted_params,
        "global": {
            "total_mutations_last_24h": total_mutations,
            "total_rollbacks_last_24h": total_rollbacks,
            "total_drift_flags_last_24h": total_drift_flags,
            "avg_stability_score": round(global_avg_stability, 2),
            "governance_health": health,
            "stability_score": global_stability
        }
    }


def intent_loop():
    while True:
        process_pending_intents()
        time.sleep(INTENT_POLL_INTERVAL_SEC)


def heartbeat_loop():
    while True:
        heartbeat_tick()
        time.sleep(HEARTBEAT_INTERVAL_SEC)


def policy_application_loop():
    while True:
        policy_application_tick()
        time.sleep(2)


def drift_detection_loop():
    while True:
        drift_detection_tick()
        time.sleep(5)


def metrics_loop():
    while True:
        metrics_tick()
        time.sleep(10) # Compute metrics every 10 seconds


if __name__ == "__main__":
    threading.Thread(target=intent_loop, daemon=True).start()
    threading.Thread(target=heartbeat_loop, daemon=True).start()
    threading.Thread(target=policy_application_loop, daemon=True).start()
    threading.Thread(target=drift_detection_loop, daemon=True).start()
    threading.Thread(target=metrics_loop, daemon=True).start()
    app.run(host="0.0.0.0", port=8080)



