package main

import (
	"fmt"
	"math/rand"
	"os"
)

// TipSelector defines the interface for a TSA
type TipSelector interface {
	TipSelect(Tx, *Sim) []int
}

// RandomWalker defines the interface for a random walk
type RandomWalker interface {
	RandomWalkStep(int, *Sim) (int, int)
	RandomWalkStepBack(int, *Sim) int
	RandomWalkStepInfinity(int, *Sim) (int, int)
}

// URTS defines a concrete type of a TSA
type URTS struct {
	TipSelector
}

// URW defines a concrete type of a TSA
type URW struct {
	TipSelector
	RandomWalker
}

// BRW defines a concrete type of a TSA
type BRW struct {
	TipSelector
	RandomWalker
}

// HPS implements the "Heaviest Pair Selection" algorithm, where the tips are selected in such a way
// that the number of referenced transactions is maximized.
type HPS struct{}

type refSet map[int]struct{}

// Compute the union of two sets.
func union(a, b refSet) refSet {
	r := make(refSet, len(a)+len(b))

	for x, y := range a {
		r[x] = y
	}
	for x, y := range b {
		r[x] = y
	}
	return r
}

// getReferences returns the set of all the transactions directly or indirectly referenced by t.
// The references are computed using recursion and dynamic programming.
func getReferences(t int, tangle []Tx, cache []refSet) refSet {
	if cache[t] != nil {
		return cache[t]
	}

	result := make(refSet)
	for _, r := range tangle[t].ref {
		result = union(result, getReferences(r, tangle, cache))
		result[r] = struct{}{}
	}
	cache[t] = result
	return result
}

func (HPS) TipSelect(t Tx, sim *Sim) []int {
	if len(sim.tips) < sim.param.K {
		return append([]int{}, sim.tips...)
	}

	if sim.param.K != 2 {
		panic("Only 2 references supported.")
	}

	cache := make([]refSet, len(sim.tangle))
	cache[0] = make(refSet) // genesis has no referenced txs

	var bestWeight int
	var bestResults [][]int

	// loop through all pairs of tips and find the pair with the most referenced txs.
	for _, t1 := range sim.tips {
		ref1 := getReferences(t1, sim.tangle, cache)

		for _, t2 := range sim.tips {
			if t1 >= t2 {
				continue // we don't care about the order in the pair
			}

			ref2 := getReferences(t2, sim.tangle, cache)
			if len(ref1)+len(ref2) < bestWeight {
				continue
			}

			refs := union(ref1, ref2)
			refs[t1] = struct{}{}
			refs[t2] = struct{}{}

			weight := len(refs)
			if weight > bestWeight {
				bestWeight = weight
				bestResults = [][]int{[]int{t1, t2}}
			} else if weight == bestWeight {
				bestResults = append(bestResults, []int{t1, t2})
			}
		}
	}

	// select a random pair from the set of best pairs
	result := bestResults[rand.Intn(len(bestResults))]
	// mark the approval time if needed
	for _, x := range result {
		if sim.tangle[x].firstApproval < 0 {
			sim.tangle[x].firstApproval = t.time
		}
	}
	return result
}

// TipSelect selects k tips
func (URTS) TipSelect(t Tx, sim *Sim) []int {
	tipsApproved := make([]int, sim.param.K)
	//var j int

	AppID := 0
	for i := 0; i < sim.param.K; i++ {
		//URTS with repetition
		if len(sim.tips) == 0 {
			fmt.Println("ERROR: No tips left")
			os.Exit(0)
		}
		j := rand.Intn(len(sim.tips))

		uniqueTip := true
		if sim.param.SingleEdgeEnabled { // consider only SingleEdge Model
			for i1 := 0; i1 < min2Int(i, len(tipsApproved)); i1++ {
				if tipsApproved[i1] == sim.tips[j] {
					uniqueTip = false
				}
			}
		}
		if uniqueTip {
			tipsApproved[AppID] = sim.tips[j] //sim.tangle[sim.vTips[j]].id
			AppID++
		} else { // tip already existed and we are in the SingleEdge Model
			tipsApproved = tipsApproved[:len(tipsApproved)-1]
		}

		//tipsApproved = append(tipsApproved, sim.tangle[sim.vTips[j]].id)
		// tipsApproved[i] = sim.tips[j]
		if sim.tangle[sim.tips[j]].firstApproval < 0 {
			sim.tangle[sim.tips[j]].firstApproval = t.time
		}
	}
	return tipsApproved
}

// TipSelect selects k tips
func (tsa URW) TipSelect(t Tx, sim *Sim) []int {
	return randomWalk(tsa, t, sim)
}

// TipSelect selects k tips
func (tsa BRW) TipSelect(t Tx, sim *Sim) []int {
	return randomWalk(tsa, t, sim)
}

//RandomWalk returns the choosen tip and its index position
func (URW) RandomWalkStep(t int, sim *Sim) (int, int) {
	directApprovers := sim.tangle[t].app
	if (len(directApprovers)) == 0 {
		return t, -1
	}
	if (len(directApprovers)) == 1 {
		return directApprovers[0], 0
	}
	j := rand.Intn(len(directApprovers))
	return directApprovers[j], j
}

//RandomWalkStepInfinity returns the choosen tip and its index position when walking over the spine Tangle
func (URW) RandomWalkStepInfinity(t int, sim *Sim) (int, int) {
	directApprovers := sim.spinePastCone[t].app
	if (len(directApprovers)) == 0 {
		return t, -1
	}
	if (len(directApprovers)) == 1 {
		return directApprovers[0], 0
	}
	j := rand.Intn(len(directApprovers))
	return directApprovers[j], j
}

//RandomWalkStepBack returns the choosen tx
func (URW) RandomWalkStepBack(t int, sim *Sim) int {
	refs := sim.tangle[t].ref
	if len(refs) == 0 {
		return t
	}
	if (len(refs)) == 1 {
		return refs[0]
	}
	j := rand.Intn(len(refs))
	return refs[j]
}

//RandomWalkStep returns the chosen tip and its index position
func (BRW) RandomWalkStep(t int, sim *Sim) (choosenTip int, approverIndx int) {
	//defer sim.b.track(runningtime("BRW"))
	directApprovers := sim.tangle[t].app
	if (len(directApprovers)) == 0 {
		return t, -1
	}
	if (len(directApprovers)) == 1 {
		return directApprovers[0], 0
	}

	nw := normalizeWeights(directApprovers, sim)
	tip, j := weightedChoose(directApprovers, nw, sim.generator, sim.b)
	return tip, j
}

//RandomWalkStepInfinity returns the chosen tip and its index position when walking over the spine Tangle
func (BRW) RandomWalkStepInfinity(t int, sim *Sim) (choosenTip int, approverIndx int) {
	//defer sim.b.track(runningtime("BRW"))
	directApprovers := sim.spinePastCone[t].app
	if (len(directApprovers)) == 0 {
		return t, -1
	}
	if (len(directApprovers)) == 1 {
		return directApprovers[0], 0
	}

	nw := normalizeWeights(directApprovers, sim)
	tip, j := weightedChoose(directApprovers, nw, sim.generator, sim.b)
	return tip, j
}

//RandomWalkStepBack returns the chosen tx
func (BRW) RandomWalkStepBack(t int, sim *Sim) (choosenTip int) {
	//defer sim.b.track(runningtime("BRW"))
	refs := sim.tangle[t].ref
	if (len(refs)) == 0 {
		return t
	}
	if (len(refs)) == 1 {
		return refs[0]
	}

	nw := normalizeWeights(refs, sim)
	tip, _ := weightedChoose(refs, nw, sim.generator, sim.b)
	return tip
}

func randomWalk(tsa RandomWalker, t Tx, sim *Sim) []int {
	defer sim.b.track(runningtime("RW"))
	tipsApproved := make([]int, sim.param.K)

	AppID := 0
	for i := 0; i < sim.param.K; i++ {
		var current int
		for current, _ = tsa.RandomWalkStep(0, sim); len(sim.tangle[current].app) > 0; current, _ = tsa.RandomWalkStep(current, sim) {
		}

		uniqueTip := true
		if sim.param.SingleEdgeEnabled { // consider only SingleEdge Model
			for i1 := 0; i1 < min2Int(i, len(tipsApproved)); i1++ {
				if tipsApproved[i1] == current {
					uniqueTip = false
				}
			}
		}
		if uniqueTip {
			tipsApproved[AppID] = current //sim.tangle[sim.vTips[j]].id
			AppID++
		} else { // tip already existed and we are in the SingleEdge Model
			tipsApproved = tipsApproved[:len(tipsApproved)-1]
		}

		//tipsApproved = append(tipsApproved, sim.tangle[sim.vTips[j]].id)
		if sim.tangle[current].firstApproval < 0 {
			sim.tangle[current].firstApproval = t.time
		}
	}
	return tipsApproved
}
