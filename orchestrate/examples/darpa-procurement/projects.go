package main

import (
	"math/rand"
)

type ClassificationLevel string

const (
	Unclassified ClassificationLevel = "UNCLASSIFIED"
	Secret       ClassificationLevel = "SECRET"
	TopSecret    ClassificationLevel = "TOP SECRET"
)

type ProjectTemplate struct {
	Name           string
	Category       string
	Classification ClassificationLevel
	Description    string
	MinComponents  int
	MaxComponents  int
}

var projectTemplates = []ProjectTemplate{
	{
		Name:           "Hypersonic Flight Control Systems",
		Category:       "Advanced Aerodynamics",
		Classification: TopSecret,
		Description:    "Development of advanced flight control systems for vehicles operating at Mach 5+ velocities with adaptive surface control and thermal management",
		MinComponents:  5,
		MaxComponents:  8,
	},
	{
		Name:           "Quantum Entanglement Communication",
		Category:       "Secure Communications",
		Classification: TopSecret,
		Description:    "Quantum-based unhackable communication system leveraging entanglement for instantaneous secure data transmission",
		MinComponents:  4,
		MaxComponents:  6,
	},
	{
		Name:           "Autonomous Underwater Vehicle Swarms",
		Category:       "Naval Systems",
		Classification: Secret,
		Description:    "Coordinated swarm of 1000+ autonomous underwater vehicles for submarine detection and tracking using distributed sensor networks",
		MinComponents:  6,
		MaxComponents:  10,
	},
	{
		Name:           "Neural Interface Brain-Computer Systems",
		Category:       "Human Performance",
		Classification: Secret,
		Description:    "Direct neural interface enabling operators to control military systems through thought with bidirectional feedback",
		MinComponents:  4,
		MaxComponents:  7,
	},
	{
		Name:           "Directed Energy Weapon Miniaturization",
		Category:       "Advanced Weapons",
		Classification: TopSecret,
		Description:    "Portable high-energy laser weapon systems with compact power generation and thermal dissipation",
		MinComponents:  5,
		MaxComponents:  8,
	},
	{
		Name:           "Self-Healing Armor Materials",
		Category:       "Advanced Materials",
		Classification: Secret,
		Description:    "Polymer-based armor with autonomous damage repair capabilities using embedded microencapsulated healing agents",
		MinComponents:  3,
		MaxComponents:  5,
	},
	{
		Name:           "Synthetic Organism Detection Biosensors",
		Category:       "Biotechnology",
		Classification: Secret,
		Description:    "Distributed biosensor network for real-time detection and identification of engineered biological threats",
		MinComponents:  4,
		MaxComponents:  7,
	},
	{
		Name:           "Space-Based Sensor Networks",
		Category:       "Space Systems",
		Classification: TopSecret,
		Description:    "Orbital constellation of advanced sensors for global surveillance, missile tracking, and space situational awareness",
		MinComponents:  6,
		MaxComponents:  9,
	},
}

var usedProjects = make(map[int]bool)

func GetRandomProject() ProjectTemplate {
	available := []int{}
	for i := range projectTemplates {
		if !usedProjects[i] {
			available = append(available, i)
		}
	}

	if len(available) == 0 {
		usedProjects = make(map[int]bool)
		available = make([]int, len(projectTemplates))
		for i := range available {
			available[i] = i
		}
	}

	idx := available[rand.Intn(len(available))]
	usedProjects[idx] = true
	return projectTemplates[idx]
}

func ResetProjects() {
	usedProjects = make(map[int]bool)
}

func (pt ProjectTemplate) ComponentCount() int {
	if pt.MinComponents == pt.MaxComponents {
		return pt.MinComponents
	}
	return pt.MinComponents + rand.Intn(pt.MaxComponents-pt.MinComponents+1)
}

func (pt ProjectTemplate) ComplexityScore() int {
	score := pt.ComponentCount() * 10

	switch pt.Classification {
	case TopSecret:
		score += 30
	case Secret:
		score += 15
	case Unclassified:
		score += 0
	}

	return score
}
