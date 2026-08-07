import sys

with open(r'D:\pqr.info\mev\swarm\atlas_api.go', 'r') as f:
    content = f.read()

broken_str = """// HandleDarkWebProxy routes an HTTP request to the local DarkWeb proxy.
// Query param "onion" is required.
func (api *AtlasAPI) HandleDarkWebProxy(w http.ResponseWriter, r *http.Request) {
	targetOnion := r.URL.Query().Get("onion")
		http.Error(w, "Failed to open local chat_in.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s: %s\\n", msg.SenderID, msg.Message)); err != nil {
		http.Error(w, "Failed to write to local chat_in.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "reply": "Mesh payload synced to chat_in.txt"})
}"""

fixed_str = """// HandleDarkWebProxy routes an HTTP request to the local DarkWeb proxy.
// Query param "onion" is required.
func (api *AtlasAPI) HandleDarkWebProxy(w http.ResponseWriter, r *http.Request) {
	targetOnion := r.URL.Query().Get("onion")
	if targetOnion == "" {
		http.Error(w, "missing 'onion' query parameter", http.StatusBadRequest)
		return
	}

	proxy, err := NewDarkWebProxy("127.0.0.1:9050")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to init proxy: %v", err), http.StatusInternalServerError)
		return
	}

	jobID, err := proxy.QueueOnionScrape(r.Context(), targetOnion)
	if err != nil {
		http.Error(w, fmt.Sprintf("onion routing failed: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "queued",
		"job_id": jobID,
	})
}

// HandleMeshSync accepts inbound payload pushes from the 5D router (originating from Nuremberg or other mesh nodes).
// It acts as the NAT-traversed drop-in point for `chat_in.txt`.
func (api *AtlasAPI) HandleMeshSync(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	var msg ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := os.OpenFile(`C:\\logs\\chat_in.txt`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open local chat_in.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s: %s\\n", msg.SenderID, msg.Message)); err != nil {
		http.Error(w, "Failed to write to local chat_in.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "reply": "Mesh payload synced to chat_in.txt"})
}"""

if broken_str in content:
    content = content.replace(broken_str, fixed_str)
    with open(r'D:\pqr.info\mev\swarm\atlas_api.go', 'w') as f:
        f.write(content)
    print("Fixed!")
else:
    print("Broken string not found!")
