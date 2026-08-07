package agents

import (
    "context"
    "fmt"
    "log"
    "time"

    "memory-graph-go/mcp"
    "memory-graph-go/gemini"
)

func RunGenomeExtractor(ctx context.Context, m *mcp.MCPClient, missionID string) {
    agentName := "genome_extractor"

    // BOOT
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{
        "status": "running",
        "started_at": time.Now().UTC().String(),
    })

    m.PublishEvent(ctx, missionID, map[string]interface{}{
        "agent": agentName,
        "type": "agent_started",
    })

    // OBSERVE
    foiaData, _ := m.GetGlobalState(ctx, "mission:"+missionID+":foia")

    // INFER
    genome := gemini.ExtractReviewerGenome(foiaData)

    // WRITE
    m.SetGlobalState(ctx, "mission:"+missionID+":genome", genome)

    m.PublishEvent(ctx, missionID, map[string]interface{}{
        "agent": agentName,
        "type": "genome_updated",
    })

    // COMPLETE
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{
        "status": "completed",
        "completed_at": time.Now().UTC().String(),
    })

    m.PublishEvent(ctx, missionID, map[string]interface{}{
        "agent": agentName,
        "type": "agent_completed",
    })

    log.Println("Genome extractor completed.")
}

func RunRiskAgent(ctx context.Context, m *mcp.MCPClient, missionID string) {
    agentName := "risk_agent"

    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{
        "status": "running",
        "started_at": time.Now().UTC().String(),
    })
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "agent_started"})

    genome, _ := m.GetGlobalState(ctx, "mission:"+missionID+":genome")
    ticket, _ := m.GetGlobalState(ctx, "mission:"+missionID+":ticket:7")

    riskOutput := gemini.EvaluateRisk(genome, ticket)

    m.SetGlobalState(ctx, "mission:"+missionID+":ticket:7", riskOutput)
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "ticket_updated", "ticket": 7})

    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{
        "status": "completed",
        "completed_at": time.Now().UTC().String(),
    })
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "agent_completed"})
    log.Println("Risk agent completed.")
}

func RunSustainabilityAgent(ctx context.Context, m *mcp.MCPClient, missionID string) {
    agentName := "sustainability_agent"
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "running"})
    genome, _ := m.GetGlobalState(ctx, "mission:"+missionID+":genome")
    ticket, _ := m.GetGlobalState(ctx, "mission:"+missionID+":ticket:12")

    sustainabilityOutput := gemini.EvaluateSustainability(genome, ticket)
    m.SetGlobalState(ctx, "mission:"+missionID+":ticket:12", sustainabilityOutput)
    
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "completed"})
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "agent_completed"})
}

func RunComplianceAgent(ctx context.Context, m *mcp.MCPClient, missionID string) {
    agentName := "compliance_agent"
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "running"})
    genome, _ := m.GetGlobalState(ctx, "mission:"+missionID+":genome")
    ticket, _ := m.GetGlobalState(ctx, "mission:"+missionID+":ticket:33")

    complianceOutput := gemini.EvaluateCompliance(genome, ticket)
    m.SetGlobalState(ctx, "mission:"+missionID+":ticket:33", complianceOutput)
    
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "completed"})
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "agent_completed"})
}

func RunReviewerAgent(ctx context.Context, m *mcp.MCPClient, missionID string) {
    agentName := "reviewer_agent"
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "running"})
    genome, _ := m.GetGlobalState(ctx, "mission:"+missionID+":genome")
    ticket, _ := m.GetGlobalState(ctx, "mission:"+missionID+":ticket:21")

    reviewerOutput := gemini.EvaluateReviewerBehavior(genome, ticket)
    m.SetGlobalState(ctx, "mission:"+missionID+":ticket:21", reviewerOutput)
    
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "completed"})
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "agent_completed"})
}

func RunSynthesisAgent(ctx context.Context, m *mcp.MCPClient, missionID string) {
    agentName := "synthesis_agent"
    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "running"})
    
    genome, _ := m.GetGlobalState(ctx, "mission:"+missionID+":genome")

    tickets := make([]interface{}, 49)
    for i := 1; i <= 49; i++ {
        t, _ := m.GetGlobalState(ctx, "mission:"+missionID+":ticket:"+fmt.Sprint(i))
        tickets[i-1] = t
    }

    finalOutput := gemini.SynthesizeProposal(genome, tickets)
    m.SetGlobalState(ctx, "mission:"+missionID+":final_output", finalOutput)

    m.SetGlobalState(ctx, "mission:"+missionID+":agents:"+agentName, map[string]interface{}{"status": "completed"})
    m.PublishEvent(ctx, missionID, map[string]interface{}{"agent": agentName, "type": "agent_completed"})
}
