package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"pqr.info/mev/governance"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	snapshotDir := "policy_snapshots"

	switch cmd {
	case "show":
		showCmd := flag.NewFlagSet("show", flag.ExitOnError)
		configPath := showCmd.String("config", "policy.json", "path to policy configuration JSON file")
		showCmd.Parse(os.Args[2:])
		cfg := loadConfig(*configPath)
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))

	case "add-lineage":
		addCmd := flag.NewFlagSet("add-lineage", flag.ExitOnError)
		configPath := addCmd.String("config", "policy.json", "path to policy configuration JSON file")
		roleFlag := addCmd.String("role", "", "agent role (e.g. Sentinel, Analyst)")
		addFlag := addCmd.String("add", "", "lineage string to authorize")
		addCmd.Parse(os.Args[2:])

		if *roleFlag == "" || *addFlag == "" {
			log.Fatal("missing required flags: -role and -add")
		}

		cfg := loadConfig(*configPath)

		// Increment version before mutation
		ver, _ := strconv.Atoi(cfg.Version)
		cfg.Version = strconv.Itoa(ver + 1)

		cfg.LineageRules[*roleFlag] = append(cfg.LineageRules[*roleFlag], *addFlag)

		// Save and Snapshot
		saveConfig(*configPath, cfg)
		_ = governance.SnapshotPolicy(cfg, snapshotDir)

		fmt.Printf("Successfully authorized lineage '%s' for role '%s'. New Policy Version: %s\n", *addFlag, *roleFlag, cfg.Version)

	case "remove-lineage":
		removeCmd := flag.NewFlagSet("remove-lineage", flag.ExitOnError)
		configPath := removeCmd.String("config", "policy.json", "path to policy configuration JSON file")
		roleFlag := removeCmd.String("role", "", "agent role (e.g. Sentinel, Analyst)")
		removeFlag := removeCmd.String("remove", "", "lineage string to de-authorize")
		removeCmd.Parse(os.Args[2:])

		if *roleFlag == "" || *removeFlag == "" {
			log.Fatal("missing required flags: -role and -remove")
		}

		cfg := loadConfig(*configPath)

		// Increment version before mutation
		ver, _ := strconv.Atoi(cfg.Version)
		cfg.Version = strconv.Itoa(ver + 1)

		list, ok := cfg.LineageRules[*roleFlag]
		if !ok {
			log.Fatalf("no rules found for role '%s'", *roleFlag)
		}
		newList := []string{}
		for _, ln := range list {
			if ln != *removeFlag {
				newList = append(newList, ln)
			}
		}
		cfg.LineageRules[*roleFlag] = newList

		// Save and Snapshot
		saveConfig(*configPath, cfg)
		_ = governance.SnapshotPolicy(cfg, snapshotDir)

		fmt.Printf("Successfully de-authorized lineage '%s' for role '%s'. New Policy Version: %s\n", *removeFlag, *roleFlag, cfg.Version)

	case "rollback":
		rollbackCmd := flag.NewFlagSet("rollback", flag.ExitOnError)
		configPath := rollbackCmd.String("config", "policy.json", "path to policy configuration JSON file")
		versionFlag := rollbackCmd.String("version", "", "target version snapshot to rollback to")
		rollbackCmd.Parse(os.Args[2:])

		if *versionFlag == "" {
			log.Fatal("missing required flag: -version")
		}

		snapPath := filepath.Join(snapshotDir, fmt.Sprintf("policy_%s.json", *versionFlag))
		snapCfg, err := governance.LoadPolicyConfig(snapPath)
		if err != nil {
			log.Fatalf("failed to load snapshot version '%s': %v", *versionFlag, err)
		}

		// Save snapshot as the active configuration
		saveConfig(*configPath, snapCfg)
		fmt.Printf("Successfully rolled back policy to Version %s.\n", *versionFlag)

	default:
		printUsage()
		log.Fatalf("unrecognized command: %s", cmd)
	}
}

func printUsage() {
	fmt.Println("Usage: governance_cli <command> [flags]")
	fmt.Println("Commands:")
	fmt.Println("  show              Print the active policy rules")
	fmt.Println("  add-lineage       Add an authorized lineage to a role")
	fmt.Println("  remove-lineage    Remove an authorized lineage from a role")
	fmt.Println("  rollback          Rollback policy to a snapshot version")
}

func loadConfig(path string) governance.PolicyConfig {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("failed to resolve absolute path: %v", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		initial := governance.PolicyConfig{
			Version: "1",
			LineageRules: map[string][]string{
				"Sentinel": {"corridor.east", "corridor.west"},
				"Analyst":  {"corridor.east", "corridor.west"},
			},
		}
		data, _ := json.MarshalIndent(initial, "", "  ")
		os.WriteFile(absPath, data, 0644)
	}
	cfg, err := governance.LoadPolicyConfig(absPath)
	if err != nil {
		log.Fatalf("failed to load policy configuration: %v", err)
	}
	return cfg
}

func saveConfig(path string, cfg governance.PolicyConfig) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("failed to resolve absolute path: %v", err)
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	err = os.WriteFile(absPath, data, 0644)
	if err != nil {
		log.Fatalf("failed to save policy: %v", err)
	}
}
