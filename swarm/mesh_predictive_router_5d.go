package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"time"
)

// Address5D defines coordinates in 5-dimensional spacetime topology: [X, Y, Z, Epoch(T), Volatility(W)]
type Address5D struct {
	X float64
	Y float64
	Z float64
	T float64
	W float64
}

// quantize24 packs a float64 into 24 bits (dropping the lower 8 bits of float32 mantissa)
func quantize24(f float64) uint32 {
	bits := math.Float32bits(float32(f))
	return bits >> 8
}

// dequantize24 unpacks 24 bits back into a float64
func dequantize24(u uint32) float64 {
	bits := u << 8
	return float64(math.Float32frombits(bits))
}

// ToIPv6 dynamically binds the 5D topological coordinates to a routable 128-bit IPv6 address.
// Uses an fc00::/8 CJDNS/Mesh prefix (8 bits) + 5 * 24-bit quantized coordinates = 128 bits.
func (a Address5D) ToIPv6() net.IP {
	ip := make(net.IP, 16)
	ip[0] = 0xfc // Mesh ULA prefix (fc00::/8)

	// Quantize each dimension to 24 bits (3 bytes)
	qX := quantize24(a.X)
	qY := quantize24(a.Y)
	qZ := quantize24(a.Z)
	qT := quantize24(a.T)
	qW := quantize24(a.W)

	ip[1] = byte(qX >> 16)
	ip[2] = byte(qX >> 8)
	ip[3] = byte(qX)

	ip[4] = byte(qY >> 16)
	ip[5] = byte(qY >> 8)
	ip[6] = byte(qY)

	ip[7] = byte(qZ >> 16)
	ip[8] = byte(qZ >> 8)
	ip[9] = byte(qZ)

	ip[10] = byte(qT >> 16)
	ip[11] = byte(qT >> 8)
	ip[12] = byte(qT)

	ip[13] = byte(qW >> 16)
	ip[14] = byte(qW >> 8)
	ip[15] = byte(qW)

	return ip
}

// Address5DFromIPv6 decodes a 128-bit IPv6 address back into 5D topological space coordinates.
func Address5DFromIPv6(ip net.IP) Address5D {
	if len(ip) == 16 {
		qX := uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
		qY := uint32(ip[4])<<16 | uint32(ip[5])<<8 | uint32(ip[6])
		qZ := uint32(ip[7])<<16 | uint32(ip[8])<<8 | uint32(ip[9])
		qT := uint32(ip[10])<<16 | uint32(ip[11])<<8 | uint32(ip[12])
		qW := uint32(ip[13])<<16 | uint32(ip[14])<<8 | uint32(ip[15])

		return Address5D{
			X: dequantize24(qX),
			Y: dequantize24(qY),
			Z: dequantize24(qZ),
			T: dequantize24(qT),
			W: dequantize24(qW),
		}
	}
	return Address5D{}
}

// Base27Encode converts the 128-bit mesh entropy of the IPv6 address to a Base-27 string.
func Base27Encode(ip net.IP) string {
	// Custom Base-27 alphabet
	const charset = "abcdefghijklmnopqrstuvwxyz_"
	
	// Convert 16 bytes into a big.Int
	
	var num big.Int
	num.SetBytes(ip)
	
	var base big.Int
	base.SetInt64(27)
	
	var zero big.Int
	var mod big.Int
	
	var result string
	for num.Cmp(&zero) > 0 {
		num.DivMod(&num, &base, &mod)
		result = string(charset[mod.Int64()]) + result
	}
	
	// Pad to 28 chars to ensure stable length for 128 bits
	for len(result) < 28 {
		result = "a" + result
	}
	
	return result
}

// Base27Decode converts a Base-27 string back into a 128-bit IPv6 address (net.IP).
func Base27Decode(s string) (net.IP, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz_"
	
	var num big.Int
	var base big.Int
	base.SetInt64(27)
	
	var val big.Int
	
	for i := 0; i < len(s); i++ {
		idx := -1
		for j := 0; j < len(charset); j++ {
			if charset[j] == s[i] {
				idx = j
				break
			}
		}
		if idx == -1 {
			return nil, fmt.Errorf("invalid base27 character: %c", s[i])
		}
		
		num.Mul(&num, &base)
		val.SetInt64(int64(idx))
		num.Add(&num, &val)
	}
	
	b := num.Bytes()
	
	// Pad front with zeros to ensure 16 bytes
	ip := make(net.IP, 16)
	if len(b) <= 16 {
		copy(ip[16-len(b):], b)
	} else {
		copy(ip, b[len(b)-16:]) // truncation if it somehow exceeds 128 bit
	}
	
	return ip, nil
}

// Distance computes Euclidean distance in 5D coordinate space.
func (a Address5D) Distance(b Address5D) float64 {
	return math.Sqrt(
		math.Pow(a.X-b.X, 2) +
			math.Pow(a.Y-b.Y, 2) +
			math.Pow(a.Z-b.Z, 2) +
			math.Pow(a.T-b.T, 2) +
			math.Pow(a.W-b.W, 2),
	)
}

// PredictiveModelNode represents a node in the mesh network offering model access.
type PredictiveModelNode struct {
	NodeID   string
	Address  Address5D
	IsActive bool
	Model    string // e.g., "gemini-3.5-flash"
	Endpoint string // e.g., "http://192.168.12.234:8100"
}

// MeshPredictiveRouter5D coordinates query routing to model providers.
type MeshPredictiveRouter5D struct {
	Nodes  []PredictiveModelNode
	replay *TimeMachineReplay
}

// NewMeshPredictiveRouter5D instantiates a new 5D routing manager.
func NewMeshPredictiveRouter5D(replay *TimeMachineReplay) *MeshPredictiveRouter5D {
	return &MeshPredictiveRouter5D{
		Nodes:  []PredictiveModelNode{},
		replay: replay,
	}
}

// RegisterNode adds a model provider node to the 5D grid map.
func (r *MeshPredictiveRouter5D) RegisterNode(node PredictiveModelNode) {
	r.Nodes = append(r.Nodes, node)
}

// RouteHint represents a BGP route hint from the Looking Glass.
type RouteHint struct {
	ASPathLength int  `json:"as_path_length"`
	Active       bool `json:"active"`
}

// QueryRouteHint queries the Cloudflare BGP Looking Glass API for route metrics.
func (r *MeshPredictiveRouter5D) QueryRouteHint(ip net.IP) (RouteHint, error) {
	client := http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:8080/bgp/api/v1/routes/%s", ip.String())
	resp, err := client.Get(url)
	if err != nil {
		return RouteHint{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RouteHint{Active: false}, nil
	}

	var hint RouteHint
	if err := json.NewDecoder(resp.Body).Decode(&hint); err != nil {
		return RouteHint{}, err
	}
	return hint, nil
}

// FindClosestNode looks up the nearest active model provider node using 5D addressing and BGP route hints.
func (r *MeshPredictiveRouter5D) FindClosestNode(sender Address5D) (*PredictiveModelNode, error) {
	var closest *PredictiveModelNode
	minDist := math.MaxFloat64

	for i := range r.Nodes {
		node := &r.Nodes[i]
		if !node.IsActive {
			continue
		}
		// Euclidean base distance
		baseDist := sender.Distance(node.Address)

		// BGP Route Hint multiplier
		multiplier := 1.0
		hint, err := r.QueryRouteHint(node.Address.ToIPv6())
		if err == nil {
			if !hint.Active {
				// Mark inactive dynamically if route withdrawn
				node.IsActive = false
				fmt.Printf("[BGP-LOOKING-GLASS] Node %s withdrawn from BGP. Marking inactive.\n", node.NodeID)
				continue
			}
			// 10% latency penalty per AS hop
			multiplier = 1.0 + (0.1 * float64(hint.ASPathLength))
		}

		effectiveDist := baseDist * multiplier

		if effectiveDist < minDist {
			minDist = effectiveDist
			closest = node
		}
	}

	if closest == nil {
		return nil, fmt.Errorf("no active model provider nodes found in 5D mesh space")
	}
	return closest, nil
}

// RouteQuery routes an LLM request from a 5D mesh address to the closest active model provider node.
func (r *MeshPredictiveRouter5D) RouteQuery(ctx context.Context, sender Address5D, query string) (string, error) {
	node, err := r.FindClosestNode(sender)
	if err != nil {
		return "", err
	}

	fmt.Printf("[5D-ROUTER] Routing query from address (%.2f,%.2f,%.2f,%.2f,%.2f) to node %s (%s) at distance %.4f\n",
		sender.X, sender.Y, sender.Z, sender.T, sender.W, node.NodeID, node.Model, sender.Distance(node.Address),
	)

	// If this node has no remote endpoint, derive it from its 5D topological address
	targetEndpoint := node.Endpoint
	if targetEndpoint == "" {
		targetEndpoint = "http://[" + node.Address.ToIPv6().String() + "]:8100"
		fmt.Printf("[5D-ROUTER] Synthesized IPv6 endpoint from topology: %s\n", targetEndpoint)
	}

	// Prepare payload for the remote node
	payload := map[string]interface{}{
		"query": query,
		"sender": map[string]float64{
			"x": sender.X, "y": sender.Y, "z": sender.Z, "t": sender.T, "w": sender.W,
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to encode inference payload: %w", err)
	}

	// Execute HTTP POST to the node's /inference/route endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", targetEndpoint+"/inference/route", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to create inference request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach node %s: %w", node.NodeID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("node %s returned error (status %d): %s", node.NodeID, resp.StatusCode, string(body))
	}

	// Assume the remote node returns JSON {"result": "..."}
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response from node %s: %w", node.NodeID, err)
	}

	return result["result"], nil
}
