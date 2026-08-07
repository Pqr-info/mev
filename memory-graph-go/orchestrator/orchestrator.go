package orchestrator

import (
    "context"
    "fmt"
    "log"

    "memory-graph-go/agents"
    "memory-graph-go/mcp"
)

type DBClient interface {
    InsertMemoryGraph(missionID string, snapshot map[string]interface{})
    CloseMission(missionID string)
}

type Orchestrator struct {
    MissionID string
    Agents    []string
    MCP       *mcp.MCPClient
    DB        DBClient
}

func NewOrchestrator(missionID string, m *mcp.MCPClient, db DBClient) *Orchestrator {
    return &Orchestrator{
        MissionID: missionID,
        Agents: []string{
            "genome_extractor",
            "risk_agent",
            "sustainability_agent",
            "compliance_agent",
            "reviewer_agent",
        },
        MCP: m,
        DB:  db,
    }
}

func (o *Orchestrator) Run(ctx context.Context, description string) {
    o.StartMission(ctx, description)
    o.SpawnAgents(ctx)
    o.MonitorEvents(ctx)
}

func (o *Orchestrator) StartMission(ctx context.Context, description string) {
    log.Println("Starting mission:", o.MissionID)
}

func (o *Orchestrator) SpawnAgents(ctx context.Context) {
    go agents.RunGenomeExtractor(ctx, o.MCP, o.MissionID)
    go agents.RunRiskAgent(ctx, o.MCP, o.MissionID)
    go agents.RunSustainabilityAgent(ctx, o.MCP, o.MissionID)
    go agents.RunComplianceAgent(ctx, o.MCP, o.MissionID)
    go agents.RunReviewerAgent(ctx, o.MCP, o.MissionID)
}

func (o *Orchestrator) MonitorEvents(ctx context.Context) {
    // This would ideally listen to Valkey events.
}

func (o *Orchestrator) HandleEvent(ctx context.Context, evt map[string]interface{}) {
    switch evt["type"] {
    case "genome_updated":
        log.Println("Genome updated — parallel agents already running.")
    case "agent_completed":
        o.CheckIfSynthesisShouldStart(ctx)
    case "synthesis_ready":
        o.ArchiveMission(ctx)
    case "mission_closed":
        log.Println("Mission closed.")
    }
}

func (o *Orchestrator) CheckIfSynthesisShouldStart(ctx context.Context) {
    completed := 0
    for _, name := range o.Agents {
        state, _ := o.MCP.GetGlobalState(ctx, "mission:"+o.MissionID+":agents:"+name)
        if state != nil && state["status"] == "completed" {
            completed++
        }
    }

    if completed == len(o.Agents) {
        go agents.RunSynthesisAgent(ctx, o.MCP, o.MissionID)
    }
}

func (o *Orchestrator) ArchiveMission(ctx context.Context) {
    snapshot := make(map[string]interface{})
    for i := 1; i <= 49; i++ {
        ticket, _ := o.MCP.GetGlobalState(ctx, fmt.Sprintf("mission:%s:ticket:%d", o.MissionID, i))
        snapshot[fmt.Sprintf("ticket_%d", i)] = ticket
    }

    if o.DB != nil {
        o.DB.InsertMemoryGraph(o.MissionID, snapshot)
        o.DB.CloseMission(o.MissionID)
    }

    o.MCP.PublishEvent(ctx, o.MissionID, map[string]interface{}{
        "agent": "orchestrator",
        "type": "mission_closed",
    })
}
