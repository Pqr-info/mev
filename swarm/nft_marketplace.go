package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// IPRights defines the legal and commercial rights associated with the NFT.
type IPRights struct {
	Rights    RightsBlock   `json:"rights"`
	Royalties RoyaltyBlock  `json:"royalties"`
	Copyright CopyrightInfo `json:"copyright"`
	License   LicenseInfo   `json:"license"`
}

type RightsBlock struct {
	CommercialUse   bool `json:"commercial_use"`
	DerivativeWorks bool `json:"derivative_works"`
	Exclusive       bool `json:"exclusive"`
	Transferable    bool `json:"transferable"`
	Revocable       bool `json:"revocable"`
}

type RoyaltyBlock struct {
	Creator         float64 `json:"creator"`
	Holder          float64 `json:"holder"`
	OnChainEnforced bool    `json:"on_chain_enforced"`
}

type CopyrightInfo struct {
	Owner        string `json:"owner"`
	TransferHash string `json:"transfer_hash"`
	Jurisdiction string `json:"jurisdiction"`
}

type LicenseInfo struct {
	Type     string `json:"type"`
	TermsURL string `json:"terms_url"`
}

// NFTMetadata represents the standard ERC-721 / Substrate JSON metadata structure.
type NFTMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	ExternalURL string   `json:"external_url,omitempty"`
	IPRights    IPRights `json:"ip_rights"`
}

type NFTMarketplace struct {
	Substrate *SubstrateClient
	ZDriveCDN string // The public CDN URL for the Z-Drive (e.g., https://my-domain.com/assets/)
}

func NewNFTMarketplace(rpcUrl, cdnUrl string) *NFTMarketplace {
	if cdnUrl == "" {
		cdnUrl = "https://mesh.local/assets/"
	}
	return &NFTMarketplace{
		Substrate: NewSubstrateClient(rpcUrl),
		ZDriveCDN: cdnUrl,
	}
}

// HandleNFTMetadata serves the JSON metadata required by wallets and explorers.
func (nm *NFTMarketplace) HandleNFTMetadata(w http.ResponseWriter, r *http.Request) {
	// Expecting /nft/metadata?id=123
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing 'id' query parameter", http.StatusBadRequest)
		return
	}

	metadata := NFTMetadata{
		Name:        fmt.Sprintf("ImageFX Asset #%s", idStr),
		Description: "A sovereign mesh-generated ImageFX asset stored securely on the Z-Drive.",
		Image:       fmt.Sprintf("%s%s.png", nm.ZDriveCDN, idStr),
		ExternalURL: fmt.Sprintf("https://mesh.local/gallery/%s", idStr),
		IPRights: IPRights{
			Rights: RightsBlock{
				CommercialUse:   true,
				DerivativeWorks: true,
				Exclusive:       false,
				Transferable:    true,
				Revocable:       false,
			},
			Royalties: RoyaltyBlock{
				Creator:         0.10,
				Holder:          0.90,
				OnChainEnforced: true,
			},
			Copyright: CopyrightInfo{
				Owner:        "5DaddrCreatorXyZ...",
				TransferHash: "0xhash_of_signed_transfer",
				Jurisdiction: "US Copyright Act (17 U.S.C.)",
			},
			License: LicenseInfo{
				Type:     "Perpetual Worldwide Non-Exclusive License",
				TermsURL: fmt.Sprintf("https://pqr.info/ip/%s", idStr),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(metadata)
}

type MintRequest struct {
	AssetID      string `json:"asset_id"`
	CollectionID uint32 `json:"collection_id"`
	OwnerAddress string `json:"owner_address"`
}

// HandleNFTMint allows a user to mint an ImageFX asset to the Substrate blockchain.
func (nm *NFTMarketplace) HandleNFTMint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Generate the metadata URI which points back to our own Metadata API
	metadataURI := fmt.Sprintf("https://mesh.local/nft/metadata?id=%s", req.AssetID)

	txHash, err := nm.Substrate.MintNFT(r.Context(), req.CollectionID, metadataURI, req.OwnerAddress)
	if err != nil {
		http.Error(w, "failed to mint on substrate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "minted",
		"tx_hash":  txHash,
		"asset_id": req.AssetID,
	})
}

// HandleNFTListings returns active marketplace listings (mocked).
func (nm *NFTMarketplace) HandleNFTListings(w http.ResponseWriter, r *http.Request) {
	listings := []map[string]interface{}{
		{
			"collection_id": 1,
			"item_id":       1001,
			"price":         500,
			"seller":        "5DaddrXyZ...",
			"asset_url":     nm.ZDriveCDN + "1001.png",
		},
		{
			"collection_id": 1,
			"item_id":       1002,
			"price":         1200,
			"seller":        "5DaddrAbC...",
			"asset_url":     nm.ZDriveCDN + "1002.png",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(listings)
}

type BuyRequest struct {
	CollectionID uint32 `json:"collection_id"`
	ItemID       uint32 `json:"item_id"`
	BuyerAddress string `json:"buyer_address"`
}

// HandleNFTBuy facilitates purchasing a listed NFT.
func (nm *NFTMarketplace) HandleNFTBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	txHash, err := nm.Substrate.BuyNFT(r.Context(), req.CollectionID, req.ItemID, req.BuyerAddress)
	if err != nil {
		http.Error(w, "failed to buy on substrate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "purchased",
		"tx_hash": txHash,
	})
}
