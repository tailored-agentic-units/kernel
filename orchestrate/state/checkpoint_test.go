package state_test

import (
	"context"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

func TestState_CheckpointMetadata(t *testing.T) {
	observer := observability.NoOpObserver{}
	s := state.New(observer)

	if s.RunID == "" {
		t.Error("Expected non-empty RunID")
	}

	if s.CheckpointNode != "" {
		t.Errorf("Expected empty CheckpointNode, got %s", s.CheckpointNode)
	}

	if s.Timestamp.IsZero() {
		t.Error("Expected non-zero Timestamp")
	}
}

func TestState_SetCheckpointNode(t *testing.T) {
	observer := observability.NoOpObserver{}
	s := state.New(observer)

	originalTime := s.Timestamp
	originalRunID := s.RunID

	time.Sleep(10 * time.Millisecond)

	s2 := s.SetCheckpointNode("node1")

	if s2.CheckpointNode != "node1" {
		t.Errorf("Expected CheckpointNode 'node1', got %s", s2.CheckpointNode)
	}

	if s2.RunID != originalRunID {
		t.Error("Expected RunID to be preserved")
	}

	if !s2.Timestamp.After(originalTime) {
		t.Error("Expected Timestamp to be updated")
	}

	if s.CheckpointNode != "" {
		t.Error("Expected original State to be unchanged (immutability)")
	}
}

func TestState_Clone_PreservesCheckpointMetadata(t *testing.T) {
	observer := observability.NoOpObserver{}
	s := state.New(observer).
		Set("key", "value").
		SetCheckpointNode("node1")

	cloned := s.Clone()

	if cloned.RunID != s.RunID {
		t.Error("Expected RunID to be preserved in clone")
	}

	if cloned.CheckpointNode != s.CheckpointNode {
		t.Error("Expected CheckpointNode to be preserved in clone")
	}

	if !cloned.Timestamp.Equal(s.Timestamp) {
		t.Error("Expected Timestamp to be preserved in clone")
	}

	val, exists := cloned.Get("key")
	if !exists || val != "value" {
		t.Error("Expected data to be preserved in clone")
	}
}

func TestState_Set_PreservesCheckpointMetadata(t *testing.T) {
	observer := observability.NoOpObserver{}
	s := state.New(observer).SetCheckpointNode("node1")

	originalRunID := s.RunID
	originalNode := s.CheckpointNode
	originalTime := s.Timestamp

	s2 := s.Set("key", "value")

	if s2.RunID != originalRunID {
		t.Error("Expected RunID to be preserved through Set")
	}

	if s2.CheckpointNode != originalNode {
		t.Error("Expected CheckpointNode to be preserved through Set")
	}

	if !s2.Timestamp.Equal(originalTime) {
		t.Error("Expected Timestamp to be preserved through Set")
	}
}

func TestState_Merge_PreservesCheckpointMetadata(t *testing.T) {
	observer := observability.NoOpObserver{}
	s1 := state.New(observer).
		Set("key1", "value1").
		SetCheckpointNode("node1")

	s2 := state.New(observer).
		Set("key2", "value2")

	merged := s1.Merge(s2)

	if merged.RunID != s1.RunID {
		t.Error("Expected RunID from first State to be preserved")
	}

	if merged.CheckpointNode != s1.CheckpointNode {
		t.Error("Expected CheckpointNode from first State to be preserved")
	}

	val1, exists1 := merged.Get("key1")
	val2, exists2 := merged.Get("key2")

	if !exists1 || val1 != "value1" {
		t.Error("Expected key1 from first State")
	}

	if !exists2 || val2 != "value2" {
		t.Error("Expected key2 from second State")
	}
}

func TestMemoryCheckpointStore_SaveAndLoad(t *testing.T) {
	store := state.NewMemoryCheckpointStore()
	observer := observability.NoOpObserver{}
	s := state.New(observer).
		Set("key", "value").
		SetCheckpointNode("node1")

	if err := store.Save(s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load(s.RunID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.RunID != s.RunID {
		t.Error("Expected loaded RunID to match")
	}

	if loaded.CheckpointNode != s.CheckpointNode {
		t.Error("Expected loaded CheckpointNode to match")
	}

	val, exists := loaded.Get("key")
	if !exists || val != "value" {
		t.Error("Expected loaded data to match")
	}
}

func TestMemoryCheckpointStore_Load_NotFound(t *testing.T) {
	store := state.NewMemoryCheckpointStore()

	_, err := store.Load("nonexistent-id")
	if err == nil {
		t.Error("Expected error when loading nonexistent checkpoint")
	}
}

func TestMemoryCheckpointStore_Delete(t *testing.T) {
	store := state.NewMemoryCheckpointStore()
	observer := observability.NoOpObserver{}
	s := state.New(observer)

	if err := store.Save(s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := store.Delete(s.RunID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Load(s.RunID)
	if err == nil {
		t.Error("Expected error when loading deleted checkpoint")
	}
}

func TestMemoryCheckpointStore_List(t *testing.T) {
	store := state.NewMemoryCheckpointStore()
	observer := observability.NoOpObserver{}

	s1 := state.New(observer)
	s2 := state.New(observer)

	if err := store.Save(s1); err != nil {
		t.Fatalf("Save s1 failed: %v", err)
	}

	if err := store.Save(s2); err != nil {
		t.Fatalf("Save s2 failed: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(list))
	}

	found1, found2 := false, false
	for _, id := range list {
		if id == s1.RunID {
			found1 = true
		}
		if id == s2.RunID {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Expected both RunIDs in list")
	}
}

func TestMemoryCheckpointStore_Overwrite(t *testing.T) {
	store := state.NewMemoryCheckpointStore()
	observer := observability.NoOpObserver{}
	s := state.New(observer).Set("key", "value1")

	if err := store.Save(s); err != nil {
		t.Fatalf("First save failed: %v", err)
	}

	s2 := s.Set("key", "value2").SetCheckpointNode("node2")

	if err := store.Save(s2); err != nil {
		t.Fatalf("Second save failed: %v", err)
	}

	loaded, err := store.Load(s.RunID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	val, _ := loaded.Get("key")
	if val != "value2" {
		t.Errorf("Expected value2 (overwritten), got %v", val)
	}

	if loaded.CheckpointNode != "node2" {
		t.Errorf("Expected node2, got %s", loaded.CheckpointNode)
	}
}

func TestCheckpointStore_Registry(t *testing.T) {
	store, err := state.GetCheckpointStore("memory")
	if err != nil {
		t.Fatalf("GetCheckpointStore failed: %v", err)
	}

	if store == nil {
		t.Error("Expected non-nil store")
	}
}

func TestCheckpointStore_Registry_NotFound(t *testing.T) {
	_, err := state.GetCheckpointStore("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent store")
	}
}

func TestCheckpointStore_RegisterCustomStore(t *testing.T) {
	customStore := state.NewMemoryCheckpointStore()
	state.RegisterCheckpointStore("custom", customStore)

	retrieved, err := state.GetCheckpointStore("custom")
	if err != nil {
		t.Fatalf("GetCheckpointStore failed: %v", err)
	}

	if retrieved != customStore {
		t.Error("Expected retrieved store to match registered store")
	}
}

func TestGraph_Checkpoint_Disabled(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 0

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("result", "node1")
	node2 := simpleNode("result", "node2")

	graph.AddNode("node1", node1)
	graph.AddNode("node2", node2)
	graph.AddEdge("node1", "node2", nil)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node2")

	observer := observability.NoOpObserver{}
	initialState := state.New(observer)

	finalState, err := graph.Execute(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if finalState.CheckpointNode == "" {
		t.Error("Expected CheckpointNode to be set even with checkpointing disabled")
	}
}

func TestGraph_Checkpoint_SaveAtInterval(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 1
	cfg.Checkpoint.Store = "memory"

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("result", "node1")
	node2 := simpleNode("result", "node2")
	node3 := simpleNode("result", "node3")

	graph.AddNode("node1", node1)
	graph.AddNode("node2", node2)
	graph.AddNode("node3", node3)
	graph.AddEdge("node1", "node2", nil)
	graph.AddEdge("node2", "node3", nil)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node3")

	observer := observability.NoOpObserver{}
	initialState := state.New(observer)
	runID := initialState.RunID

	_, err = graph.Execute(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	store, _ := state.GetCheckpointStore("memory")
	_, err = store.Load(runID)
	if err == nil {
		t.Error("Expected checkpoint to be deleted after successful completion (Preserve=false)")
	}
}

func TestGraph_Checkpoint_PreserveOnSuccess(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 1
	cfg.Checkpoint.Store = "memory"
	cfg.Checkpoint.Preserve = true

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("result", "node1")
	node2 := simpleNode("result", "node2")

	graph.AddNode("node1", node1)
	graph.AddNode("node2", node2)
	graph.AddEdge("node1", "node2", nil)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node2")

	observer := observability.NoOpObserver{}
	initialState := state.New(observer)
	runID := initialState.RunID

	_, err = graph.Execute(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	store, _ := state.GetCheckpointStore("memory")
	loaded, err := store.Load(runID)
	if err != nil {
		t.Errorf("Expected checkpoint to be preserved, got error: %v", err)
	}

	if loaded.CheckpointNode != "node2" {
		t.Errorf("Expected final checkpoint at node2, got %s", loaded.CheckpointNode)
	}

	store.Delete(runID)
}

func TestGraph_Resume_FromCheckpoint(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 1
	cfg.Checkpoint.Store = "memory"
	cfg.Checkpoint.Preserve = true

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("step", "1")
	node2 := simpleNode("step", "2")
	node3 := simpleNode("step", "3")

	graph.AddNode("node1", node1)
	graph.AddNode("node2", node2)
	graph.AddNode("node3", node3)
	graph.AddEdge("node1", "node2", nil)
	graph.AddEdge("node2", "node3", nil)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node3")

	observer := observability.NoOpObserver{}
	initialState := state.New(observer)
	runID := initialState.RunID

	partialState, err := graph.Execute(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if partialState.CheckpointNode != "node3" {
		t.Errorf("Expected checkpoint at node3, got %s", partialState.CheckpointNode)
	}

	store, _ := state.GetCheckpointStore("memory")
	checkpointed := partialState.SetCheckpointNode("node1")
	store.Save(checkpointed)

	resumedState, err := graph.Resume(context.Background(), runID)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	if resumedState.RunID != runID {
		t.Error("Expected RunID to be preserved through resume")
	}

	val, exists := resumedState.Get("step")
	if !exists {
		t.Error("Expected step key to exist")
	}

	if val != "3" {
		t.Errorf("Expected final step value '3', got %v", val)
	}

	store.Delete(runID)
}

func TestGraph_Resume_CheckpointingDisabled(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 0

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("result", "value")
	graph.AddNode("node1", node1)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node1")

	_, err = graph.Resume(context.Background(), "any-id")
	if err == nil {
		t.Error("Expected error when resuming with checkpointing disabled")
	}
}

func TestGraph_Resume_CheckpointNotFound(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 1
	cfg.Checkpoint.Store = "memory"

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("result", "value")
	graph.AddNode("node1", node1)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node1")

	_, err = graph.Resume(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("Expected error when resuming from nonexistent checkpoint")
	}
}

func TestGraph_Resume_AtExitPoint(t *testing.T) {
	cfg := config.DefaultGraphConfig("test")
	cfg.Checkpoint.Interval = 1
	cfg.Checkpoint.Store = "memory"

	graph, err := state.NewGraph(cfg)
	if err != nil {
		t.Fatalf("NewGraph failed: %v", err)
	}

	node1 := simpleNode("result", "value")
	graph.AddNode("node1", node1)
	graph.SetEntryPoint("node1")
	graph.SetExitPoint("node1")

	observer := observability.NoOpObserver{}
	initialState := state.New(observer)
	runID := initialState.RunID

	_, err = graph.Execute(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	store, _ := state.GetCheckpointStore("memory")
	checkpointed := initialState.SetCheckpointNode("node1")
	store.Save(checkpointed)

	_, err = graph.Resume(context.Background(), runID)
	if err == nil {
		t.Error("Expected error when resuming from exit point checkpoint")
	}

	store.Delete(runID)
}

func TestState_Checkpoint_Method(t *testing.T) {
	store := state.NewMemoryCheckpointStore()
	observer := observability.NoOpObserver{}
	s := state.New(observer).
		Set("key", "value").
		SetCheckpointNode("node1")

	if err := s.Checkpoint(store); err != nil {
		t.Fatalf("Checkpoint method failed: %v", err)
	}

	loaded, err := store.Load(s.RunID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.RunID != s.RunID {
		t.Error("Expected loaded state to match")
	}
}

func simpleNode(key, value string) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s.Set(key, value), nil
	})
}
